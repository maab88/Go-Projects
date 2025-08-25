package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Simple extension -> category mapping.
// You can expand this over time; unknowns fall into "Other".
var extToCategory = map[string]string{
	// Images
	".jpg": "Images", ".jpeg": "Images", ".png": "Images", ".gif": "Images",
	".webp": "Images", ".bmp": "Images", ".tiff": "Images", ".heic": "Images",

	// Video
	".mp4": "Video", ".mov": "Video", ".mkv": "Video", ".avi": "Video",
	".wmv": "Video", ".flv": "Video", ".webm": "Video",

	// Audio
	".mp3": "Audio", ".wav": "Audio", ".aac": "Audio", ".flac": "Audio",
	".m4a": "Audio", ".ogg": "Audio",

	// Docs
	".pdf": "Docs", ".doc": "Docs", ".docx": "Docs", ".xls": "Docs",
	".xlsx": "Docs", ".ppt": "Docs", ".pptx": "Docs",
	".txt": "Docs", ".rtf": "Docs", ".md": "Docs", ".csv": "Docs",

	// Archives
	".zip": "Archives", ".rar": "Archives", ".7z": "Archives", ".gz": "Archives",
	".tar": "Archives",

	// Code
	".go": "Code", ".cs": "Code", ".js": "Code", ".ts": "Code", ".jsx": "Code",
	".tsx": "Code", ".py": "Code", ".java": "Code", ".rb": "Code",
	".php": "Code", ".c": "Code", ".cpp": "Code", ".h": "Code", ".hpp": "Code",
}

type job struct {
	srcPath string
	info    os.FileInfo
}

type result struct {
	srcPath string
	dstPath string
	err     error
	action  string // "move" or "skip"
}

// Move record for manifest/undo
type Move struct {
	Src  string    `json:"src"`
	Dst  string    `json:"dst"`
	When time.Time `json:"when"`
}

func main() {
	var (
		srcDir        string
		dstDir        string
		dryRun        bool
		workers       int
		includeHidden bool
		undoManifest  string
	)

	flag.StringVar(&srcDir, "src", ".", "Source directory to organize")
	flag.StringVar(&dstDir, "dest", "", "Destination root directory (default: same as src)")
	flag.BoolVar(&dryRun, "dry-run", false, "Print actions without making changes")
	flag.IntVar(&workers, "workers", 8, "Number of worker goroutines")
	flag.BoolVar(&includeHidden, "include-hidden", false, "Include hidden files (.* on Unix)")
	flag.StringVar(&undoManifest, "undo", "", "Undo using the given manifest JSON and exit")
	flag.Parse()

	if dstDir == "" {
		dstDir = srcDir
	}

	// Validate source; we allow undo to run without a dest check.
	mustBeDir(srcDir)

	// Undo mode short-circuit
	if undoManifest != "" {
		if err := undoFromManifest(undoManifest, dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "undo failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Normal run: validate dest too
	mustBeDir(dstDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Producer/consumer channels
	jobs := make(chan job, 256)
	results := make(chan result, 256)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(ctx, &wg, jobs, results, dstDir, dryRun)
	}

	// Walk in a separate goroutine so we can consume results concurrently
	go func() {
		defer close(jobs)
		_ = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				results <- result{srcPath: path, err: err}
				return nil
			}
			// Skip directories
			if info.IsDir() {
				return nil
			}
			// Optional: skip hidden files
			if !includeHidden && isHidden(info) {
				return nil
			}
			// Skip files already under a categorized subfolder of dst
			// (prevents re-moving if src==dst)
			if inCategorizedSubfolder(dstDir, path) {
				return nil
			}

			jobs <- job{srcPath: path, info: info}
			return nil
		})
	}()

	// Collector: close results when workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	var moved, skipped, failed int
	var moves []Move
	start := time.Now()

	for r := range results {
		if r.err != nil {
			failed++
			fmt.Printf("ERROR  %s -> %s  (%v)\n", r.srcPath, r.dstPath, r.err)
			continue
		}
		switch r.action {
		case "move":
			moved++
			if dryRun {
				fmt.Printf("DRYRUN %s -> %s\n", r.srcPath, r.dstPath)
			} else {
				fmt.Printf("MOVED  %s -> %s\n", r.srcPath, r.dstPath)
				moves = append(moves, Move{Src: r.srcPath, Dst: r.dstPath, When: time.Now()})
			}
		case "skip":
			skipped++
			fmt.Printf("SKIP   %s\n", r.srcPath)
		default:
			// no-op
		}
	}

	// Write manifest for undo
	if !dryRun && len(moves) > 0 {
		if mf, err := writeManifest(dstDir, moves); err != nil {
			fmt.Printf("WARN   failed to write manifest: %v\n", err)
		} else {
			fmt.Printf("Manifest saved: %s\n", mf)
		}
	}

	elapsed := time.Since(start).Truncate(time.Millisecond)
	fmt.Printf("\nDone in %s | moved=%d skipped=%d failed=%d\n", elapsed, moved, skipped, failed)
}

func worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	jobs <-chan job,
	results chan<- result,
	dstRoot string,
	dryRun bool,
) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case j, ok := <-jobs:
			if !ok {
				return
			}
			r := handleJob(j, dstRoot, dryRun)
			results <- r
		}
	}
}

func handleJob(j job, dstRoot string, dryRun bool) result {
	ext := strings.ToLower(filepath.Ext(j.info.Name()))
	category, ok := extToCategory[ext]
	if !ok || category == "" {
		category = "Other"
	}

	// Compute destination folder and filename
	dstDir := filepath.Join(dstRoot, category)
	dstPath := filepath.Join(dstDir, j.info.Name())

	// If the source and destination are the same path, skip
	if sameFile(j.srcPath, dstPath) {
		return result{srcPath: j.srcPath, dstPath: dstPath, action: "skip"}
	}

	// Ensure destination directory exists
	if !dryRun {
		if err := os.MkdirAll(dstDir, 0o755); err != nil {
			return result{srcPath: j.srcPath, dstPath: dstPath, err: err}
		}
	}

	// Resolve name conflicts
	if !dryRun {
		if exists(dstPath) {
			var err error
			dstPath, err = nextAvailableName(dstPath)
			if err != nil {
				return result{srcPath: j.srcPath, dstPath: dstPath, err: err}
			}
		}
	}

	// Move (or simulate)
	if dryRun {
		return result{srcPath: j.srcPath, dstPath: dstPath, action: "move"}
	}

	if err := moveFile(j.srcPath, dstPath); err != nil {
		return result{srcPath: j.srcPath, dstPath: dstPath, err: err}
	}
	return result{srcPath: j.srcPath, dstPath: dstPath, action: "move"}
}

func mustBeDir(path string) {
	info, err := os.Stat(path)
	if err != nil {
		exitf("path %q: %v", path, err)
	}
	if !info.IsDir() {
		exitf("%q is not a directory", path)
	}
}

func exitf(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

// Returns true if path is inside dstRoot/<any-known-category>/...
func inCategorizedSubfolder(dstRoot, path string) bool {
	absRoot, _ := filepath.Abs(dstRoot)
	absPath, _ := filepath.Abs(path)

	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return false
	}

	first := strings.Split(rel, string(filepath.Separator))[0]
	if first == "." || first == "" {
		return false
	}

	for _, cat := range []string{"Images", "Video", "Audio", "Docs", "Archives", "Code", "Other"} {
		if first == cat {
			return true
		}
	}
	return false
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func sameFile(a, b string) bool {
	absA, _ := filepath.Abs(a)
	absB, _ := filepath.Abs(b)
	return absA == absB
}

func nextAvailableName(path string) (string, error) {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	for i := 1; i < 10_000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", name, i, ext))
		if !exists(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("too many name conflicts for %q", path)
}

// moveFile attempts a fast rename; if crossing devices, it falls back to copy+remove.
func moveFile(src, dst string) error {
	// try rename
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// fallback: copy then remove
	if err := copyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(in *os.File) {
		_ = in.Close()
	}(in)

	// Create with same perms as source when possible
	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

// isHidden: on Unix, files starting with '.'; on Windows, this is a naive check.
// (You can improve Windows detection using syscall attributes in a future iteration.)
func isHidden(info os.FileInfo) bool {
	name := info.Name()
	return strings.HasPrefix(name, ".")
}

// ----- Manifest & Undo -----

func writeManifest(dstRoot string, moves []Move) (string, error) {
	if len(moves) == 0 {
		return "", nil
	}
	dir := filepath.Join(dstRoot, ".organizer-manifests")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	name := fmt.Sprintf("moves-%s.json", time.Now().Format("20060102-150405"))
	path := filepath.Join(dir, name)

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(moves); err != nil {
		return "", err
	}
	return path, nil
}

func undoFromManifest(manifest string, dryRun bool) error {
	f, err := os.Open(manifest)
	if err != nil {
		return err
	}
	defer f.Close()

	var moves []Move
	if err := json.NewDecoder(f).Decode(&moves); err != nil {
		return err
	}

	var undone, skipped, failed int
	// Reverse order to safely unwind nested moves
	for i := len(moves) - 1; i >= 0; i-- {
		m := moves[i]
		if !exists(m.Dst) {
			fmt.Printf("SKIP   missing: %s (already moved/deleted)\n", m.Dst)
			skipped++
			continue
		}
		target := m.Src
		if exists(target) {
			// Don’t clobber anything that reappeared at the original location
			var err error
			target, err = nextAvailableName(target)
			if err != nil {
				fmt.Printf("ERROR  undo %s -> %s (%v)\n", m.Dst, target, err)
				failed++
				continue
			}
		}
		if dryRun {
			fmt.Printf("DRYRUN UNDO %s -> %s\n", m.Dst, target)
			undone++
			continue
		}
		if err := moveFile(m.Dst, target); err != nil {
			fmt.Printf("ERROR  undo %s -> %s (%v)\n", m.Dst, target, err)
			failed++
			continue
		}
		fmt.Printf("UNDONE %s -> %s\n", m.Dst, target)
		undone++
	}
	fmt.Printf("\nUndo summary: undone=%d skipped=%d failed=%d\n", undone, skipped, failed)
	// Try to remove empty category dirs in the directory that contained the manifest.
	if !dryRun {
		// derive dstRoot from manifest path
		if root := filepath.Dir(filepath.Dir(manifest)); root != "" {
			removeEmptyCategoryDirs(root)
		}
	}
	return nil
}

// removeEmptyCategoryDirs deletes empty category folders under dstRoot.
// Safe: only removes the known category directories if they are empty.
func removeEmptyCategoryDirs(dstRoot string) {
	cats := []string{"Images", "Video", "Audio", "Docs", "Archives", "Code", "Other"}
	for _, c := range cats {
		dir := filepath.Join(dstRoot, c)
		// Only attempt if the dir exists
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			// Check emptiness
			empty := true
			_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if p != dir {
					empty = false
					return fmt.Errorf("stop")
				} // early stop
				return nil
			})
			if empty {
				_ = os.Remove(dir) // remove empty dir
			}
		}
	}
}

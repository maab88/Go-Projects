package utils

import (
	"fmt"
	"time"
)

type CustomDate time.Time

func (cd *CustomDate) Scan(src any) error {
	switch v := src.(type) {
	case time.Time:
		*cd = CustomDate(v)
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into CustomDate", v)
	}
}

func (cd CustomDate) Time() time.Time {
	return time.Time(cd)
}

func (cd CustomDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + cd.Time().Format("2006-01-02") + `"`), nil
}

func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	parsed, err := time.Parse(`"2006-01-02"`, string(data))
	if err != nil {
		return err
	}
	*cd = CustomDate(parsed)
	return nil
}

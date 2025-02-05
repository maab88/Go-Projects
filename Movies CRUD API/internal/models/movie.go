package models

import "MoviesCRUDAPI/pkg/utils"

type Movie struct {
	ID          int              `json:"id"`
	Title       string           `json:"title"`
	ReleaseDate utils.CustomDate `json:"release_date"`
	DirectorID  int              `json:"director_id"`
}

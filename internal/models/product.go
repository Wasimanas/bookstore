package models

import ( 
    "time"
)


type Product struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
    Price  float64`json:"price"`
	Author string `json:"author"`
	Year   int    `json:"year"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

}


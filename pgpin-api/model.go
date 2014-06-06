package main

import (
	"time"
)

type pinSlim struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type pin struct {
	Id              string     `json:"id"`
	Name            string     `json:"name"`
	DbId            string     `json:"db_id"`
	Query           string     `json:"query"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	QueryStartedAt  *time.Time `json:"query_started_at"`
	QueryFinishedAt *time.Time `json:"query_finished_at"`
	ResultsFields   NullJson   `json:"results_fields"`
	ResultsRows     NullJson   `json:"results_rows"`
	ResultsError    *string    `json:"results_error"`
	DeletedAt       *time.Time `json:"-"`
}

type dbSlim struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type db struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	Url       string     `json:"url"`
	AddedAt   time.Time  `json:"added_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	RemovedAt *time.Time `json:"-"`
}

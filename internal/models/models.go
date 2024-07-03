package models

import "time"

type User struct {
	ID             int    `json:"id" example:"1"`
	PassportNumber string `json:"passportNumber" example:"AB1234567"`
	Surname        string `json:"surname" example:"Smith"`
	Name           string `json:"name" example:"John"`
	Patronymic     string `json:"patronymic" example:"Michael"`
	Address        string `json:"address" example:"123 Main St, City"`
}

type Workload struct {
	TaskID  int `json:"taskId" example:"1"`
	Hours   int `json:"hours" example:"8"`
	Minutes int `json:"minutes" example:"30"`
}

type Task struct {
	ID        int       `json:"id" example:"1"`
	UserID    int       `json:"userId" example:"1"`
	StartTime time.Time `json:"startTime" example:"2023-07-03T09:00:00Z"`
	EndTime   time.Time `json:"endTime" example:"2023-07-03T17:00:00Z"`
}

type People struct {
	Surname    string `json:"surname" example:"Smith"`
	Name       string `json:"name" example:"John"`
	Patronymic string `json:"patronymic" example:"Michael"`
	Address    string `json:"address" example:"123 Main St, City"`
}
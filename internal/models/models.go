package models

import "time"

type User struct {
	ID             int    `json:"id"`
	PassportNumber string `json:"passportNumber"`
	Surname        string `json:"surname"`
	Name           string `json:"name"`
	Patronymic     string `json:"patronymic"`
	Address        string `json:"address"`
}

type Workload struct {
	TaskID int           `json:"taskId"`
	Hours  int           `json:"hours"`
	Minutes int          `json:"minutes"`
}

type Task struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type People struct {
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
	Address    string `json:"address"`
}
package models

import "time"

type User struct {
	ID             int       `json:"id" example:"1"`
	PassportNumber string    `json:"passportNumber" example:"1234 5678"`
	Surname        string    `json:"surname" example:"Smith"`
	Name           string    `json:"name" example:"John"`
	Patronymic     string    `json:"patronymic" example:"Michael"`
	Address        string    `json:"address" example:"123 Main St, City"`
	CreatedAt      time.Time `json:"createdAt" example:"2023-07-03"`
	UpdatedAt      time.Time `json:"updatedAt" example:"2023-07-03"`
}

type Workload struct {
	TaskID      int    `json:"taskId" example:"1"`
	Description string `json:"description" example:"Project planning"`
	Hours       int    `json:"hours" example:"8"`
	Minutes     int    `json:"minutes" example:"30"`
}

type Task struct {
	ID          int       `json:"id" example:"1"`
	UserID      int       `json:"userId" example:"1"`
	Description string    `json:"description" example:"Project planning"`
	StartTime   time.Time `json:"startTime" example:"2023-07-03"`
	EndTime     time.Time `json:"endTime,omitempty" example:"2023-07-03"`
	CreatedAt   time.Time `json:"createdAt" example:"2023-07-03"`
}

type TimeEntry struct {
	ID        int            `json:"id" example:"1"`
	TaskID    int            `json:"taskId" example:"1"`
	StartTime time.Time      `json:"startTime" example:"2023-07-03"`
	EndTime   time.Time     `json:"endTime,omitempty" example:"2023-07-03"`
	Duration  time.Duration `json:"duration,omitempty" example:"8h30m"`
	CreatedAt time.Time      `json:"createdAt" example:"2023-07-03"`
}

type People struct {
	Surname    string `json:"surname" example:"Smith"`
	Name       string `json:"name" example:"John"`
	Patronymic string `json:"patronymic" example:"Michael"`
	Address    string `json:"address" example:"123 Main St, City"`
}

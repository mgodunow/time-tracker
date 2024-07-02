package repository

import (
	"database/sql"
	"fmt"
	"time"
	"timeTracker/internal/models"
)

type Repository interface {
	AddUser(user models.User) (models.User, error)
	GetUsers(page, limit int, filters map[string]string) ([]models.User, error)
	GetUserWorkload(userID int, start, end time.Time) ([]models.Workload, error)
	StartUserTask(userID, taskID int) (models.Task, error)
	StopUserTask(userID, taskID int) (models.Task, error)
	DeleteUser(id int) error
	UpdateUser(user models.User) (models.User, error)
	User(id int) (models.User, error)
}

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(host, port, user, password, dbname string) *postgresRepo {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return &postgresRepo{db: db}
}

func (p *postgresRepo) AddUser(user models.User) (models.User, error) {
	panic("implement me!")
}
func (p *postgresRepo) GetUsers(page, limit int, filters map[string]string) ([]models.User, error) {
	panic("implement me!")
}
func (p *postgresRepo) GetUserWorkload(userID int, start, end time.Time) ([]models.Workload, error) {
	panic("implement me!")
}
func (p *postgresRepo) StartUserTask(userID, taskID int) (models.Task, error) {
	panic("implement me!")
}
func (p *postgresRepo) StopUserTask(userID, taskID int) (models.Task, error) {
	panic("implement me!")
}
func (p *postgresRepo) DeleteUser(id int) error {
	panic("implement me!")
}
func (p *postgresRepo) UpdateUser(user models.User) (models.User, error) {
	panic("implement me!")
}
func (p *postgresRepo) User(id int) (models.User, error) {
	panic("implement me!")
}

func NewRepository(host, port, user, password, dbname string) Repository {
	return NewPostgresRepo(host, port, user, password, dbname)
}

package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"timeTracker/internal/models"

	_ "github.com/lib/pq"
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
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return &postgresRepo{db: db}
}

func (p *postgresRepo) AddUser(user models.User) (models.User, error) {
	query := `
		INSERT INTO users (passport_number, surname, name, patronymic, address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := p.db.QueryRow(query, user.PassportNumber, user.Surname, user.Name, user.Patronymic, user.Address).Scan(&user.ID)
	if err != nil {
		return user, fmt.Errorf("error adding user to database: %w", err)
	}

	return user, nil
}

// TODO: making page & limit optional
func (p *postgresRepo) GetUsers(page, limit int, filters map[string]string) ([]models.User, error) {
	query := `SELECT id, passport_number, surname, name, patronymic, address FROM users WHERE 1=1`

	var whereParams []interface{}
	paramCounter := 1

	for field, value := range filters {
		if value != "" {
			query += fmt.Sprintf(" AND %s ILIKE $%d", field, paramCounter)
			whereParams = append(whereParams, value+"%")
			paramCounter++
		}
	}

	query += fmt.Sprintf(" ORDER BY id LIMIT $%d OFFSET $%d", paramCounter, paramCounter+1)
	offset := (page - 1) * limit
	whereParams = append(whereParams, limit, offset)

	stmt, err := p.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(whereParams...)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.PassportNumber, &u.Surname, &u.Name, &u.Patronymic, &u.Address); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
func (p *postgresRepo) GetUserWorkload(userID int, start, end time.Time) ([]models.Workload, error) {
	query := `
	SELECT task_id, 
		   EXTRACT(HOUR FROM SUM(end_time - start_time)) AS hours,
		   EXTRACT(MINUTE FROM SUM(end_time - start_time)) AS minutes
	FROM tasks
	WHERE user_id = $1 AND start_time >= $2 AND end_time <= $3
	GROUP BY task_id
	ORDER BY SUM(end_time - start_time) DESC`

	rows, err := p.db.Query(query, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workloads []models.Workload
	for rows.Next() {
		var w models.Workload
		if err := rows.Scan(&w.TaskID, &w.Hours, &w.Minutes); err != nil {
			return nil, err
		}
		workloads = append(workloads, w)
	}

	return workloads, nil
}
func (p *postgresRepo) StartUserTask(userID, taskID int) (models.Task, error) {
	query := `
	INSERT INTO tasks (user_id, task_id, start_time)
	VALUES ($1, $2, $3)
	RETURNING id, user_id, task_id, start_time`

	var task models.Task
	err := p.db.QueryRow(query, userID, taskID, time.Now()).Scan(&task.ID, &task.UserID, &taskID, &task.StartTime)
	if err != nil {
		return task, err
	}

	return task, nil
}
func (p *postgresRepo) StopUserTask(userID, taskID int) (models.Task, error) {
	query := `
		UPDATE tasks
		SET end_time = $1
		WHERE user_id = $2 AND task_id = $3 AND end_time IS NULL
		RETURNING id, user_id, task_id, start_time, end_time`

	var task models.Task
	err := p.db.QueryRow(query, time.Now(), userID, taskID).Scan(&task.ID, &task.UserID, &taskID, &task.StartTime, &task.EndTime)
	if err != nil {
		return task, err
	}

	return task, nil
}
func (p *postgresRepo) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := p.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
func (p *postgresRepo) UpdateUser(user models.User) (models.User, error) {
	query := `
		UPDATE users
		SET passport_number = $1, surname = $2, name = $3, patronymic = $4, address = $5
		WHERE id = $6
		RETURNING id, passport_number, surname, name, patronymic, address`

	err := p.db.QueryRow(query, user.PassportNumber, user.Surname, user.Name, user.Patronymic, user.Address, user.ID).
		Scan(&user.ID, &user.PassportNumber, &user.Surname, &user.Name, &user.Patronymic, &user.Address)
	if err != nil {
		return user, err
	}

	return user, nil
}
func (p *postgresRepo) User(id int) (models.User, error) {
	query := `
		SELECT id, passport_number, surname, name, patronymic, address
		FROM users
		WHERE id = $1`

	var user models.User
	err := p.db.QueryRow(query, id).Scan(&user.ID, &user.PassportNumber, &user.Surname, &user.Name, &user.Patronymic, &user.Address)
	if err != nil {
		return user, err
	}

	return user, nil
}

func NewRepository(host, port, user, password, dbname string) Repository {
	return NewPostgresRepo(host, port, user, password, dbname)
}

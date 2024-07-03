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
	SELECT t.id, t.description, 
		   ROUND(EXTRACT(EPOCH FROM SUM(te.duration))/3600)::integer AS hours,
		   ROUND(MOD(EXTRACT(EPOCH FROM SUM(te.duration))::integer, 3600)/60)::integer AS minutes
	FROM tasks t
	JOIN time_entries te ON t.id = te.task_id
	WHERE t.user_id = $1 AND te.start_time >= $2 AND te.end_time <= $3
	GROUP BY t.id, t.description
	ORDER BY SUM(te.duration) DESC`

	rows, err := p.db.Query(query, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workloads []models.Workload
	for rows.Next() {
		var w models.Workload
		if err := rows.Scan(&w.TaskID, &w.Description, &w.Hours, &w.Minutes); err != nil {
			return nil, err
		}
		workloads = append(workloads, w)
	}

	return workloads, nil
}
func (p *postgresRepo) StartUserTask(userID, taskID int) (models.Task, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return models.Task{}, err
	}
	defer tx.Rollback()

	var task models.Task
	taskQuery := `
    SELECT id, user_id, description, created_at
    FROM tasks
    WHERE id = $1 AND user_id = $2`

	err = tx.QueryRow(taskQuery, taskID, userID).
		Scan(&task.ID, &task.UserID, &task.Description, &task.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Task{}, errors.New("task not found or doesn't belong to the user")
		}
		return models.Task{}, err
	}

	var activeEntries int
	checkActiveQuery := `
    SELECT COUNT(*)
    FROM time_entries
    WHERE task_id = $1 AND end_time IS NULL`

	err = tx.QueryRow(checkActiveQuery, taskID).Scan(&activeEntries)
	if err != nil {
		return models.Task{}, err
	}
	if activeEntries > 0 {
		return models.Task{}, errors.New("task is already active")
	}

	timeEntryQuery := `
    INSERT INTO time_entries (task_id, start_time)
    VALUES ($1, $2)
    RETURNING id, start_time`

	var timeEntryID int
	var startTime time.Time
	err = tx.QueryRow(timeEntryQuery, taskID, time.Now()).Scan(&timeEntryID, &startTime)
	if err != nil {
		return models.Task{}, err
	}

	task.StartTime = startTime

	if err = tx.Commit(); err != nil {
		return models.Task{}, err
	}

	return task, nil
}
func (p *postgresRepo) StopUserTask(userID, taskID int) (models.Task, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return models.Task{}, err
	}
	defer tx.Rollback()

	query := `
		UPDATE time_entries
		SET end_time = $1, duration = $1 - start_time
		WHERE task_id = $2 AND end_time IS NULL
		RETURNING id, task_id, start_time, end_time, duration`

	var timeEntry models.TimeEntry
	var durationStr string
	err = tx.QueryRow(query, time.Now(), taskID).
		Scan(&timeEntry.ID, &timeEntry.TaskID, &timeEntry.StartTime, &timeEntry.EndTime, &durationStr)
	if err != nil {
		return models.Task{}, err
	}
	duration, err := parseDuration(durationStr)
	if err != nil {
		return models.Task{}, err
	}
	timeEntry.Duration = duration

	taskQuery := `
		SELECT id, user_id, description, created_at
		FROM tasks
		WHERE id = $1`

	var task models.Task
	err = tx.QueryRow(taskQuery, taskID).
		Scan(&task.ID, &task.UserID, &task.Description, &task.CreatedAt)
	if err != nil {
		return models.Task{}, err
	}

	task.StartTime = timeEntry.StartTime
	task.EndTime = timeEntry.EndTime

	if err = tx.Commit(); err != nil {
		return models.Task{}, err
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

func parseDuration(s string) (time.Duration, error) {
	var hours, minutes, seconds, microseconds int

	_, err := fmt.Sscanf(s, "%d:%d:%d.%d", &hours, &minutes, &seconds, &microseconds)
	if err != nil {
		return 0, err
	}

	duration := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second +
		time.Duration(microseconds)*time.Microsecond

	return duration, nil
}

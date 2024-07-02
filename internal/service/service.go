package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"timeTracker/internal/models"
	"timeTracker/internal/repository"
)

type UserService struct {
	repo                repository.Repository
	GetByPassportDomain string
}

func NewUserService(repo repository.Repository, getByPassportDomain string) *UserService {
	return &UserService{
		repo:                repo,
		GetByPassportDomain: getByPassportDomain,
	}
}

func (s *UserService) AddUser(user models.User) (models.User, error) {
	passportParts := strings.Split(user.PassportNumber, " ")
	if len(passportParts) != 2 {
		return user, fmt.Errorf("invalid passport number format")
	}

	resp, err := http.Get(fmt.Sprintf("%s?passportSerie=%s&passportNumber=%s", s.GetByPassportDomain, passportParts[0], passportParts[1]))
	if err != nil {
		return user, fmt.Errorf("error querying getByPassport API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return user, fmt.Errorf("getByPassport API returned not OK status: %d", resp.StatusCode)
	}

	var peopleInfo models.People
	if err := json.NewDecoder(resp.Body).Decode(&peopleInfo); err != nil {
		return user, fmt.Errorf("error decoding external API response: %w", err)
	}

	user.Surname = peopleInfo.Surname
	user.Name = peopleInfo.Name
	user.Patronymic = peopleInfo.Patronymic
	user.Address = peopleInfo.Address

	enrichedUser, err := s.repo.AddUser(user)
	if err != nil {
		return user, fmt.Errorf("error saving user to database: %w", err)
	}

	return enrichedUser, nil
}

func (s *UserService) GetUsers(page, limit int, filters map[string]string) ([]models.User, error) {
	return s.repo.GetUsers(page, limit, filters)
}

func (s *UserService) GetUserWorkload(userID int, start, end time.Time) ([]models.Workload, error) {
	return s.repo.GetUserWorkload(userID, start, end)
}

func (s *UserService) StartUserTask(userID, taskID int) (models.Task, error) {
	return s.repo.StartUserTask(userID, taskID)
}

func (s *UserService) StopUserTask(userID, taskID int) (models.Task, error) {
	return s.repo.StopUserTask(userID, taskID)
}

func (s *UserService) DeleteUser(id int) error {
	return s.repo.DeleteUser(id)
}

func (s *UserService) UpdateUser(user models.User) (models.User, error) {
	existingUser, err := s.repo.User(user.ID)
	if err != nil {
		return user, fmt.Errorf("error getting existing user: %w", err)
	}

	if user.PassportNumber != "" {
		existingUser.PassportNumber = user.PassportNumber
	}
	if user.Surname != "" {
		existingUser.Surname = user.Surname
	}
	if user.Name != "" {
		existingUser.Name = user.Name
	}
	if user.Patronymic != "" {
		existingUser.Patronymic = user.Patronymic
	}
	if user.Address != "" {
		existingUser.Address = user.Address
	}

	updatedUser, err := s.repo.UpdateUser(existingUser)
	if err != nil {
		return user, fmt.Errorf("error updating user in database: %w", err)
	}

	return updatedUser, nil
}

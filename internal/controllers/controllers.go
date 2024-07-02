package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"timeTracker/internal/models"
	"timeTracker/internal/service"

	"github.com/gorilla/mux"
)

const (
	InternalServerErrorMessage = "internal server error"
	BadRequestMessage          = "bad request"
)

type Handler struct {
	userService *service.UserService
	logger      *slog.Logger
}

func NewHandler(userService *service.UserService, logger *slog.Logger) *Handler {
	return &Handler{userService: userService, logger: logger}
}

func (h *Handler) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/users", h.Users).Methods("GET")
	r.HandleFunc("/users/{id}/workload", h.GetUserWorkload).Methods("GET")
	r.HandleFunc("/users/{id}/tasks/{taskId}/start", h.StartUserTask).Methods("POST")
	r.HandleFunc("/users/{id}/tasks/{taskId}/stop", h.StopUserTask).Methods("POST")
	r.HandleFunc("/users/{id}", h.DeleteUser).Methods("DELETE")
	r.HandleFunc("/users/{id}", h.UpdateUser).Methods("PUT")
	r.HandleFunc("/users", h.AddUser).Methods("POST")

	return r
}

func (h *Handler) Users(w http.ResponseWriter, r *http.Request) {
	const op = "controller GetUsers: "
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}

	filters := filters(r)

	users, err := h.userService.GetUsers(page, limit, filters)
	if err != nil {
		h.logger.With("operation: ", op).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(users); err != nil {
		h.logger.With("operation: ", op).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}
	h.logger.Debug(fmt.Sprintf("return all users with page=%d limit=%d", page, limit))
}

func filters(r *http.Request) map[string]string {
	surname := r.URL.Query().Get("surname")
	name := r.URL.Query().Get("name")
	passport_number := r.URL.Query().Get("passport_number")
	patronymic := r.URL.Query().Get("patronymic")
	address := r.URL.Query().Get("address")

	return map[string]string{"surname": surname, "name": name,
		"passport_number": passport_number, "patronymic": patronymic,
		"address": address}
}

func (h *Handler) GetUserWorkload(w http.ResponseWriter, r *http.Request) {
	const op = "controller GetUserWorkLoad: "
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}
	start, err := time.Parse("2006-01-02", r.URL.Query().Get("start"))
	if err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}
	end, err := time.Parse("2006-01-02", r.URL.Query().Get("end"))
	if err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}

	workload, err := h.userService.GetUserWorkload(id, start, end)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			h.logger.With("id", id,
				"start", start,
				"end", end).Error(err.Error())
			http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err = json.NewEncoder(w).Encode(workload); err != nil {
		h.logger.With("id", id,
			"start", start,
			"end", end).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
	}

	h.logger.With("userID", id).Debug("return user's workload")
}

func (h *Handler) StartUserTask(w http.ResponseWriter, r *http.Request) {
	const op = "controller StartUserTask: "
	userId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}
	taskId, err := strconv.Atoi(mux.Vars(r)["taskId"])
	if err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}

	task, err := h.userService.StartUserTask(userId, taskId)
	if err != nil {
		h.logger.With(
			"userID", userId,
			"taskID", taskId,
		).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(task); err != nil {
		h.logger.With(
			"userID", userId,
			"taskID", taskId,
		).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
	}

	h.logger.With("userID", userId,
		"taskID", taskId).Debug("started user's task")
}

func (h *Handler) StopUserTask(w http.ResponseWriter, r *http.Request) {
	const op = "controller StopUserTask: "
	userId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}
	taskId, err := strconv.Atoi(mux.Vars(r)["taskId"])
	if err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}

	task, err := h.userService.StopUserTask(userId, taskId)
	if err != nil {
		h.logger.With("operation: ", op,
			"taskID", task.ID,
			"userID", task.UserID).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err = json.NewEncoder(w).Encode(task); err != nil {
		h.logger.With("operation: ", op,
			"taskID", task.ID,
			"userID", task.UserID).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
	}

	h.logger.With("userID", userId,
		"taskID", taskId).Debug("stoped user's task")
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	const op = "controller DeleteUser: "
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}

	err = h.userService.DeleteUser(id)
	if err != nil {
		h.logger.With("userID", id).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	h.logger.With("userID", id).Debug("deleted user")

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	const op = "controller UpdateUser: "
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}

	user.ID = id
	updatedUser, err := h.userService.UpdateUser(user)
	if err != nil {
		h.logger.With("userID", id).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err = json.NewEncoder(w).Encode(updatedUser); err != nil {
		h.logger.With("userID", id).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}
	h.logger.With("userID", id).Debug("updated user")
}

func (h *Handler) AddUser(w http.ResponseWriter, r *http.Request) {
	const op = "controller AddUser: "
	var newUser models.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		h.logger.With("operation: ", op).Info(err.Error())
		http.Error(w, BadRequestMessage, http.StatusBadRequest)
		return
	}

	enrichedUser, err := h.userService.AddUser(newUser)
	if err != nil {
		h.logger.With("userID", newUser.ID).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(enrichedUser); err != nil {
		h.logger.With("userID", newUser.ID).Error(err.Error())
		http.Error(w, InternalServerErrorMessage, http.StatusInternalServerError)
		return
	}
	h.logger.With("userID", enrichedUser.ID).Debug("created user")
}

package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/models"
	"github.com/Fserlut/gophermart/internal/services"
)

type Handler struct {
	logger *slog.Logger

	// TODO тут не лучше использовать интерфейс?
	userService *services.UserService
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegisterOrLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Error on decode request", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req.Password == "" || req.Login == "" {
		h.logger.Error("Bad request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	cookie, err := h.userService.Register(req)
	if err != nil {
		if errors.Is(err, &lib.ErrUserExists{}) {
			h.logger.Error("User already exist")
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}

		h.logger.Error("error from service: ", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegisterOrLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Error on decode request", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req.Password == "" || req.Login == "" {
		h.logger.Error("Bad request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	cookie, err := h.userService.Login(req)

	if err != nil {
		if errors.Is(err, &lib.ErrWrongPasswordOrLogin{}) {
			h.logger.Error("Wrong password")
			http.Error(w, "Wrong login or password", http.StatusBadRequest)
			return
		} else if errors.Is(err, &lib.ErrUserNotFound{}) {
			h.logger.Error("User not found")
			http.Error(w, "User not registered", http.StatusBadRequest)
			return
		}
		h.logger.Error("error from service: ", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func NewHandler(log *slog.Logger, userService *services.UserService) *Handler {
	return &Handler{
		logger:      log,
		userService: userService,
	}
}

package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/logger"
	user2 "github.com/Fserlut/gophermart/internal/models/user"
	"github.com/Fserlut/gophermart/internal/services/user"
)

type UserHandler struct {
	logger      logger.Logger
	userService *user.ServiceUser
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req user2.RegisterOrLoginRequest

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
	w.Header().Set("Authorization", cookie.Value)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req user2.RegisterOrLoginRequest

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
	w.Header().Set("Authorization", cookie.Value)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func NewUserHandler(log logger.Logger, userService *user.ServiceUser) *UserHandler {
	h := &UserHandler{
		logger:      log,
		userService: userService,
	}

	return h
}

package handlers

import (
	"encoding/json"
	"errors"
	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/models"
	"github.com/Fserlut/gophermart/internal/services"
	"io"
	"log/slog"
	"net/http"
)

type Handler struct {
	logger *slog.Logger

	// TODO тут не лучше использовать интерфейс?
	userService  *services.UserService
	orderService *services.OrderService
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
	w.Header().Set("Authorization", cookie.Value)
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
	w.Header().Set("Authorization", cookie.Value)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderNumber := string(body)

	if !lib.CheckLuhn(orderNumber) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	code, _ := h.orderService.CreateOrder(r.Context(), orderNumber)

	w.WriteHeader(code)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.orderService.GetOrdersByUserID(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func NewHandler(log *slog.Logger, userService *services.UserService, orderService *services.OrderService) *Handler {
	return &Handler{
		logger:       log,
		userService:  userService,
		orderService: orderService,
	}
}

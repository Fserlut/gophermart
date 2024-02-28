package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/models"
	"github.com/Fserlut/gophermart/internal/services/order"
	"github.com/Fserlut/gophermart/internal/services/user"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	logger *slog.Logger

	// TODO тут не лучше использовать интерфейс?
	userService   *user.ServiceUser
	orderService  *order.ServiceOrder
	ordersChannel chan string
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

	if code == http.StatusAccepted {
		h.ordersChannel <- orderNumber
	}

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

func (h *Handler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	balance, err := h.orderService.GetUserBalance(r.Context())

	if err != nil {
		h.logger.Error("Error from order service", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req models.WithdrawRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Error on decode request", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !lib.CheckLuhn(req.Order) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	code, err := h.orderService.Withdraw(r.Context(), req)

	if err != nil {
		h.logger.Error("Error from order service", err.Error())
		w.WriteHeader(code)
		return
	}

	w.WriteHeader(code)
}

func (h *Handler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	res, err := h.orderService.Withdrawals(r.Context())

	if err != nil {
		h.logger.Error("Error from order service", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func NewHandler(log *slog.Logger, userService *user.ServiceUser, orderService *order.ServiceOrder) *Handler {

	h := &Handler{
		logger:        log,
		userService:   userService,
		orderService:  orderService,
		ordersChannel: make(chan string, 100),
	}

	//go func() {
	//	for orderNumber := range h.ordersChannel {
	//		err := h.orderService.UpdateOrderStatus(orderNumber)
	//		if err != nil {
	//			h.logger.Error(fmt.Sprintf("Error on delete url %s", err.Error()))
	//		}
	//	}
	//}()

	go func() {
		for orderNumber := range h.ordersChannel {
		Retry:
			err := h.orderService.UpdateOrderStatus(orderNumber)
			if err != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, "rate limit exceeded") {
					h.logger.Error("Rate limit exceeded, retrying after 1 minute")
					time.Sleep(1 * time.Minute) // Ожидание перед повторной попыткой
					goto Retry
				} else if strings.Contains(errMsg, "not finished") {
					h.logger.Error("Order not finished, retrying later")
					go func() { h.ordersChannel <- orderNumber }()
					continue
				} else {
					h.logger.Error(fmt.Sprintf("Error getting order info: %s", errMsg))
				}
			}
		}
	}()

	return h
}

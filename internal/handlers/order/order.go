package order

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/logger"
	user2 "github.com/Fserlut/gophermart/internal/models/user"
	"github.com/Fserlut/gophermart/internal/services/order"
)

type OrderHandler struct {
	logger       logger.Logger
	orderService *order.ServiceOrder
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
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

	err = h.orderService.CreateOrder(r.Context(), orderNumber)

	if err != nil {
		if errors.Is(err, &lib.ErrOrderAlreadyCreated{}) {
			w.WriteHeader(http.StatusOK)
			return
		} else if errors.Is(err, &lib.ErrOrderAlreadyCreatedByOtherUser{}) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
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

func (h *OrderHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	balance, err := h.orderService.GetUserBalance(r.Context())

	if err != nil {
		h.logger.Error("Error from order service", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

func (h *OrderHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req user2.WithdrawRequest

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

func (h *OrderHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	res, err := h.orderService.Withdrawals(r.Context())

	if err != nil {
		h.logger.Error("Error from order service", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func NewOrderHandler(log logger.Logger, orderService *order.ServiceOrder) *OrderHandler {
	return &OrderHandler{
		logger:       log,
		orderService: orderService,
	}
}

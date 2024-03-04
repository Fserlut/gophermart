package order

import (
	"encoding/json"
	"fmt"
	"github.com/Fserlut/gophermart/internal/lib"
	user2 "github.com/Fserlut/gophermart/internal/models/user"
	"github.com/Fserlut/gophermart/internal/services/order"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type OrderHandler struct {
	logger        *slog.Logger
	orderService  *order.ServiceOrder
	ordersChannel chan string
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

	code, _ := h.orderService.CreateOrder(r.Context(), orderNumber)

	if code == http.StatusAccepted {
		h.ordersChannel <- orderNumber
	}

	w.WriteHeader(code)
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

func NewOrderHandler(log *slog.Logger, orderService *order.ServiceOrder) *OrderHandler {

	h := &OrderHandler{
		logger:        log,
		orderService:  orderService,
		ordersChannel: make(chan string, 100),
	}

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

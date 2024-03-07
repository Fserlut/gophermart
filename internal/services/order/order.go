package order

import (
	"context"
	"errors"
	"net/http"

	"github.com/Fserlut/gophermart/internal/config"
	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/logger"
	"github.com/Fserlut/gophermart/internal/models/order"
	"github.com/Fserlut/gophermart/internal/models/user"
)

type ServiceOrder struct {
	logger          logger.Logger
	cfg             *config.Config
	orderRepository orderRepository
	processor       Processor
}

type orderRepository interface {
	GetOrderByNumber(string) (*order.Order, error)
	CreateOrder(orderNumber string, UserUUID string, withdraw *float64) error
	GetOrdersByUserID(string) ([]order.Order, error)
	GetUserBalance(string) (*user.BalanceResponse, error)
	Withdrawals(string) ([]user.WithdrawalsResponse, error)
	UpdateOrder(orderNumber string, status string, accrual *float64) error
}

type Processor interface {
	Process(orderNumber string)
}

func (o ServiceOrder) CreateOrder(ctx context.Context, orderNumber string) error {
	userID, ok := ctx.Value(lib.UserContextKey).(string)
	if !ok || userID == "" {
		return &lib.NotFoundUserIDInContext{}
	}

	err := o.orderRepository.CreateOrder(orderNumber, userID, nil)

	if err != nil {
		return err
	}

	o.processor.Process(orderNumber)

	return nil
}

func (o ServiceOrder) GetOrdersByUserID(ctx context.Context) ([]order.Order, error) {
	userID, ok := ctx.Value(lib.UserContextKey).(string)
	if !ok || userID == "" {
		return nil, &lib.NotFoundUserIDInContext{}
	}
	orders, err := o.orderRepository.GetOrdersByUserID(userID)

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (o ServiceOrder) GetUserBalance(ctx context.Context) (*user.BalanceResponse, error) {
	userID, ok := ctx.Value(lib.UserContextKey).(string)
	if !ok || userID == "" {
		return nil, &lib.NotFoundUserIDInContext{}
	}

	balance, err := o.orderRepository.GetUserBalance(userID)

	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (o ServiceOrder) Withdraw(ctx context.Context, toWithdraw user.WithdrawRequest) (int, error) {
	userID, ok := ctx.Value(lib.UserContextKey).(string)
	if !ok || userID == "" {
		return http.StatusUnauthorized, &lib.NotFoundUserIDInContext{}
	}

	balance, err := o.orderRepository.GetUserBalance(userID)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	availableBalance := balance.Current - balance.Withdrawn

	sum := float64(toWithdraw.Sum)

	if sum > availableBalance {
		return http.StatusPaymentRequired, errors.New("")
	}

	err = o.orderRepository.CreateOrder(toWithdraw.Order, userID, &sum)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (o ServiceOrder) Withdrawals(ctx context.Context) ([]user.WithdrawalsResponse, error) {
	userID, ok := ctx.Value(lib.UserContextKey).(string)
	if !ok || userID == "" {
		return nil, &lib.NotFoundUserIDInContext{}
	}

	res, err := o.orderRepository.Withdrawals(userID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func NewOrderService(log logger.Logger, cfg *config.Config, orderRepository orderRepository, processor Processor) *ServiceOrder {
	return &ServiceOrder{
		logger:          log,
		cfg:             cfg,
		orderRepository: orderRepository,
		processor:       processor,
	}
}

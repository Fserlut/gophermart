package order

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Fserlut/gophermart/internal/config"
	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/models/order"
	"github.com/Fserlut/gophermart/internal/models/user"
)

type ServiceOrder struct {
	logger          *slog.Logger
	orderRepository orderRepository
	cfg             *config.Config
}

type orderRepository interface {
	GetOrderByNumber(string) (*order.Order, error)
	CreateOrder(orderNumber string, UserUUID string, withdraw *float64) error
	GetOrdersByUserID(string) ([]order.Order, error)
	GetUserBalance(string) (*user.BalanceResponse, error)
	Withdrawals(string) ([]user.WithdrawalsResponse, error)
	Update(orderNumber string, status string, accrual *float64) error
}

func (o ServiceOrder) CreateOrder(ctx context.Context, orderNumber string) (int, error) {
	userID, ok := ctx.Value(lib.UserContextKey).(string)
	fmt.Println(userID, ok)
	if !ok || userID == "" {
		return http.StatusUnauthorized, &lib.NotFoundUserIDInContext{}
	}

	err := o.orderRepository.CreateOrder(orderNumber, userID, nil)

	if err != nil {
		if errors.Is(err, &lib.ErrOrderAlreadyCreated{}) {
			return http.StatusOK, nil
		} else if errors.Is(err, &lib.ErrOrderAlreadyCreatedByOtherUser{}) {
			return http.StatusConflict, nil
		}

		return http.StatusInternalServerError, err
	}

	return http.StatusAccepted, nil
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

func (o ServiceOrder) UpdateOrderStatus(orderNumber string) error {
	orderResult, err := lib.GetOrderInfo(fmt.Sprintf("%s/api/orders/%s", o.cfg.AccrualSystemAddress, orderNumber))

	if err != nil {
		return err
	}

	if orderResult.Status == "INVALID" || orderResult.Status == "PROCESSED" {
		err = o.orderRepository.Update(orderResult.Order, orderResult.Status, orderResult.Accrual)
		if err != nil {
			return err
		}
		return nil
	}

	if orderResult.Status == "REGISTERED" {
		return fmt.Errorf("not finished")
	}

	return nil
}

func NewOrderService(log *slog.Logger, orderRepository orderRepository, cfg *config.Config) *ServiceOrder {
	return &ServiceOrder{
		logger:          log,
		orderRepository: orderRepository,
		cfg:             cfg,
	}
}

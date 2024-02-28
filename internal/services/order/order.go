package order

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fserlut/gophermart/internal/config"
	"net/http"

	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/models"
)

type ServiceOrder struct {
	orderRepository orderRepository
	cfg             *config.Config
}

type orderRepository interface {
	GetOrderByNumber(string) (*models.Order, error)
	CreateOrder(orderNumber string, userUUID string, withdraw *float64) error
	GetOrdersByUserID(string) ([]models.Order, error)
	GetUserBalance(string) (*models.UserBalanceResponse, error)
	Withdrawals(string) ([]models.WithdrawalsResponse, error)
	Update(orderNumber string, status string, accrual *float64) error
}

func (o ServiceOrder) CreateOrder(ctx context.Context, orderNumber string) (int, error) {
	//TODO нормально ли userID передавать через context?
	if userID, ok := ctx.Value("userID").(string); ok && userID != "" {
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
	//TODO нормально ли возвращать код ответа?
	return http.StatusUnauthorized, errors.New("unauthorized")
}

func (o ServiceOrder) GetOrdersByUserID(ctx context.Context) ([]models.Order, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, &lib.NotFoundUserIDInContext{}
	}
	orders, err := o.orderRepository.GetOrdersByUserID(userID)

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (o ServiceOrder) GetUserBalance(ctx context.Context) (*models.UserBalanceResponse, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, &lib.NotFoundUserIDInContext{}
	}

	balance, err := o.orderRepository.GetUserBalance(userID)

	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (o ServiceOrder) Withdraw(ctx context.Context, toWithdraw models.WithdrawRequest) (int, error) {
	userID, ok := ctx.Value("userID").(string)
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

func (o ServiceOrder) Withdrawals(ctx context.Context) ([]models.WithdrawalsResponse, error) {
	userID, ok := ctx.Value("userID").(string)
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
	order, err := lib.GetOrderInfo(fmt.Sprintf("%s/api/orders/%s", o.cfg.AccrualSystemAddress, orderNumber))

	if err != nil {
		return err
	}

	if order.Status == "INVALID" || order.Status == "PROCESSED" {
		err = o.orderRepository.Update(order.Order, order.Status, order.Accrual)
		if err != nil {
			return err
		}
		return nil
	}

	if order.Status == "REGISTERED" {
		return errors.New("not finished")
	}

	return nil
}

func NewOrderService(orderRepository orderRepository, cfg *config.Config) *ServiceOrder {
	return &ServiceOrder{
		orderRepository: orderRepository,
		cfg:             cfg,
	}
}

package order

import (
	"context"
	"errors"
	"net/http"

	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/models"
)

type ServiceOrder struct {
	orderRepository orderRepository
}

type orderRepository interface {
	GetOrderByNumber(string) (*models.Order, error)
	CreateOrder(orderNumber string, userUUID string, withdraw *float64) error
	GetOrdersByUserID(string) ([]models.Order, error)
	GetUserBalance(string) (*models.UserBalanceResponse, error)
}

func (o ServiceOrder) CreateOrder(ctx context.Context, orderNumber string) (int, error) {
	//TODO нормально ли userID передавать через context?
	if userID, ok := ctx.Value("userID").(string); ok && userID != "" {
		err := o.orderRepository.CreateOrder(orderNumber, userID, nil)

		//TODO Тут должна быть какая-то горутина

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

func NewOrderService(orderRepository orderRepository) *ServiceOrder {
	return &ServiceOrder{
		orderRepository: orderRepository,
	}
}

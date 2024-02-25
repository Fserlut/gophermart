package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/models"
)

type OrderService struct {
	orderRepository orderRepository
}

type orderRepository interface {
	GetOrderByNumber(string) (*models.Order, error)
	CreateOrder(orderNumber string, userUUID string) error
	GetOrdersByUserID(string) ([]models.Order, error)
}

func (o OrderService) CreateOrder(ctx context.Context, orderNumber string) (int, error) {
	//TODO нормально ли userID передавать через context?
	if userID, ok := ctx.Value("userID").(string); ok && userID != "" {
		err := o.orderRepository.CreateOrder(orderNumber, userID)

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

func (o OrderService) GetOrdersByUserID(ctx context.Context) ([]models.Order, error) {
	if userID, ok := ctx.Value("userID").(string); ok && userID != "" {
		orders, err := o.orderRepository.GetOrdersByUserID(userID)

		if err != nil {
			return make([]models.Order, 0), err
		}

		return orders, nil
	}

	return make([]models.Order, 0), &lib.NotFoundUserIDInContext{}
}

func NewOrderService(orderRepository orderRepository) *OrderService {
	return &OrderService{
		orderRepository: orderRepository,
	}
}

package accrual

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Fserlut/gophermart/internal/config"
	"github.com/Fserlut/gophermart/internal/lib"
	"github.com/Fserlut/gophermart/internal/logger"
)

type Accrual struct {
	logger         logger.Logger
	cfg            *config.Config
	ctx            context.Context
	cancelFunc     context.CancelFunc
	orderProcessor OrderProcessor
	ordersChannel  chan string
}

type OrderProcessor interface {
	UpdateOrder(orderNumber string, status string, accrual *float64) error
}

func (a *Accrual) Run() {
	go func() {
		for {
			select {
			case <-a.ctx.Done():
				a.logger.Info("Stopping accrual processing")
				return
			case orderNumber, ok := <-a.ordersChannel:
				if !ok {
					a.logger.Info("Channel was closed")
					return
				}

				fmt.Println("Processing order:", orderNumber)
				a.processOrderWithRetries(a.ctx, orderNumber)
			}
		}
	}()
}

func (a *Accrual) processOrderWithRetries(ctx context.Context, orderNumber string) {
	for {
		orderResult, err := lib.GetOrderInfo(fmt.Sprintf("%s/api/orders/%s", a.cfg.AccrualSystemAddress, orderNumber))
		if err == nil && (orderResult.Status == "PROCESSED" || orderResult.Status == "INVALID") {
			err = a.orderProcessor.UpdateOrder(orderNumber, orderResult.Status, orderResult.Accrual)
			if err != nil {
				a.logger.Error(fmt.Sprintf("Error updating order status in system: %s", err.Error()))
			}
			return
		} else if err == nil && orderResult.Status == "REGISTERED" {
			if !a.waitForRetry(ctx, 10*time.Second) {
				return
			}
			go func() { a.ordersChannel <- orderNumber }()
			return
		} else if errors.Is(err, &lib.TooManyRequestsError{}) {
			a.logger.Error("Rate limit exceeded, retrying after 1 minute")
			if !a.waitForRetry(ctx, time.Minute) {
				return
			}
		} else {
			a.logger.Error(fmt.Sprintf("Error getting order info: %s", err.Error()))
			return
		}
	}
}

func (a *Accrual) waitForRetry(ctx context.Context, delay time.Duration) bool {
	select {
	case <-time.After(delay):
		return true
	case <-ctx.Done():
		return false
	}
}

func (a *Accrual) Process(orderNumber string) {
	a.ordersChannel <- orderNumber
}

func NewAccrualService(ctx context.Context, log logger.Logger, cfg *config.Config, orderProcessor OrderProcessor) *Accrual {
	c, cancel := context.WithCancel(ctx)

	return &Accrual{
		logger:         log,
		cfg:            cfg,
		ordersChannel:  make(chan string, 100),
		ctx:            c,
		cancelFunc:     cancel,
		orderProcessor: orderProcessor,
	}
}

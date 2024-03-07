package lib

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Fserlut/gophermart/internal/models/order"
)

func GetOrderInfo(link string) (*order.OrderInfo, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var info order.OrderInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			return nil, fmt.Errorf("decoding response: %w", err)
		}
		return &info, nil
	case http.StatusNoContent:
		return nil, &OrderNotFound{}
	case http.StatusTooManyRequests:
		return nil, &TooManyRequestsError{}
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("internal server error")
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

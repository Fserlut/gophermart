package lib

import (
	"encoding/json"
	"fmt"
	"github.com/Fserlut/gophermart/internal/models"
	"net/http"
)

func GetOrderInfo(link string) (*models.OrderInfo, error) {
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
		var info models.OrderInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			return nil, fmt.Errorf("decoding response: %w", err)
		}
		return &info, nil
	case http.StatusNoContent:
		return nil, fmt.Errorf("order not found")
	case http.StatusTooManyRequests:
		return nil, fmt.Errorf("rate limit exceeded")
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("internal server error")
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AnxVit/go-musthave-diploma-tpl/internal/model"
	"github.com/AnxVit/go-musthave-diploma-tpl/internal/utils"
	"github.com/avast/retry-go/v5"
	"github.com/go-resty/resty/v2"
)

type AccuralClient struct {
	client *http.Client

	retrier *retry.Retrier
	addr    string
}

func NewAccuralClient(ctx context.Context, addr string) *AccuralClient {
	retrier := utils.NewRetryer(
		3,
		time.Duration(1)*time.Second,
		time.Duration(5)*time.Second,
		func(err error) bool {
			if httpErr, ok := err.(interface{ StatusCode() int }); ok {
				code := httpErr.StatusCode()
				return code == 429
			}

			return false
		},
	)
	agent := &AccuralClient{
		client: &http.Client{
			Timeout: 100 * time.Millisecond,
		},

		retrier: retrier,

		addr: addr,
	}

	return agent
}

func (a *AccuralClient) GetResult(orderNumber string) (*model.AccuralResponse, error) {
	client := resty.New()
	var (
		rawBody []byte
		status  int
	)

	err := a.retrier.Do(func() error {
		httpResp, err := client.R().
			SetHeader("Content-Length", "0").
			Get(a.addr + "/api/orders/" + orderNumber)

		if err != nil {
			return err
		}

		if httpResp.StatusCode() != 204 && httpResp.StatusCode() != 200 {
			return &CodeError{
				err:  httpResp.String(),
				code: httpResp.StatusCode(),
			}
		}

		rawBody = httpResp.Body()
		status = httpResp.StatusCode()
		return nil
	})

	if err != nil {
		return nil, err
	}

	if status == 200 {
		var resp model.AccuralResponse
		if err := json.Unmarshal(rawBody, &resp); err != nil {
			return nil, err
		}

		return &resp, nil
	}

	return nil, nil
}

func (a *AccuralClient) Register(orderNumber string) error {
	client := resty.New()

	req := model.AccuralRequest{
		Order: orderNumber,
	}

	bytesBody, _ := json.Marshal(req)

	err := a.retrier.Do(func() error {
		httpResp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(bytesBody).
			Post(a.addr + "/api/orders")

		if err != nil {
			return err
		}

		if httpResp.StatusCode() >= 300 && httpResp.StatusCode() != 409 {
			return &CodeError{
				err:  httpResp.String(),
				code: httpResp.StatusCode(),
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

type CodeError struct {
	err  string
	code int
}

func (c *CodeError) Error() string {
	return fmt.Sprintf("%s (%d)", c.err, c.code)
}

func (c *CodeError) StatusCode() int {
	return c.code
}

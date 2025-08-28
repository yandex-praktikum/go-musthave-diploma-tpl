package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	"github.com/kdv2001/loyalty/internal/domain"
	"github.com/kdv2001/loyalty/internal/pkg/logger"
	"github.com/kdv2001/loyalty/internal/pkg/serviceerrors"
)

const retryNums = 3

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	client    httpClient
	serverURL url.URL
}

func NewClient(client httpClient, serverURL url.URL) *Client {
	return &Client{
		client:    client,
		serverURL: serverURL,
	}
}

type accrualsResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func (c *Client) GetAccruals(ctx context.Context, orderID domain.ID) (*domain.Order, error) {
	var err error
	var res *domain.Order
	logger.Infof(ctx, "start get accrual")
	for i := 0; i <= retryNums; i++ {
		res, err = c.getAccruals(ctx, orderID)
		if err != nil {
			sErr := &serviceerrors.AppError{}
			if errors.As(err, &sErr); sErr.Code == http.StatusTooManyRequests {
				continue
			}

			return nil, err
		}

		break
	}

	logger.Infof(ctx, "finish get accrual %v", res)
	return res, nil
}

func (c *Client) getAccruals(ctx context.Context, orderID domain.ID) (*domain.Order, error) {
	getAccrualsURL := c.serverURL.JoinPath("api/orders", strconv.FormatUint(orderID.ID, 10))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getAccrualsURL.String(), nil)
	if err != nil {
		return nil, serviceerrors.NewAppError(err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, serviceerrors.NewAppError(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, serviceerrors.NewAppError(err)
	}

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNoContent:
			return nil, serviceerrors.NewNoContent().Wrap(domain.ErrNoContent, "")
		case http.StatusInternalServerError:
			return nil, serviceerrors.NewAppError(nil)
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		val := resp.Header.Get("Retry-After")
		parseVal, pErr := strconv.ParseInt(val, 10, 64)
		if pErr != nil {
			return nil, serviceerrors.NewAppError(pErr)
		}
		tick := time.NewTicker(time.Duration(parseVal) * time.Second)
		select {
		case <-tick.C:
			return nil, serviceerrors.NewTooManyRequests()
		case <-ctx.Done():
			return nil, nil
		}
	}

	ar := new(accrualsResponse)
	err = json.Unmarshal(body, ar)
	if err != nil {
		return nil, serviceerrors.NewAppError(err)
	}

	return &domain.Order{
		ID:    orderID,
		State: statusToDomain(ar.Status),
		AccrualAmount: domain.Money{
			Currency: string(domain.GopherMarketBonuses),
			Amount:   decimal.NewFromFloat(ar.Accrual),
		},
	}, nil
}

func statusToDomain(state string) domain.AccrualState {
	switch state {
	case "REGISTERED":
		return domain.Processing
	case "PROCESSING":
		return domain.Processing
	case "PROCESSED":
		return domain.Processed
	}
	return domain.Invalid
}

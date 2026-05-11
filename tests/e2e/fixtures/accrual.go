package fixtures

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	AccrualPort = "7091"
)

type AccrualSuite struct {
	AccrualContainer testcontainers.Container
	AccrualURL       string
}

// SetupAccrual запускает контейнер accrual
func (s *AccrualSuite) SetupAccrual(ctx context.Context) error {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:       "../../..", // Путь от tests/e2e/fixtures к корню проекта
			Dockerfile:    "docker/accrual/Dockerfile",
			PrintBuildLog: true,
		},
		ExposedPorts: []string{AccrualPort + "/tcp"},
		// используем ожидание порта
		WaitingFor: wait.ForListeningPort(AccrualPort + "/tcp").WithStartupTimeout(60 * time.Second),
		Env: map[string]string{
			"RUN_ADDRESS": "0.0.0.0:" + AccrualPort,
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return fmt.Errorf("failed to start accrual container: %w", err)
	}

	s.AccrualContainer = container

	host, err := container.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get host: %w", err)
	}

	port, err := container.MappedPort(ctx, AccrualPort)
	if err != nil {
		return fmt.Errorf("failed to get port: %w", err)
	}

	s.AccrualURL = fmt.Sprintf("http://%s:%s", host, port.Port())
	return nil
}

// TearDownAccrual останавливает контейнер
func (s *AccrualSuite) TearDownAccrual(ctx context.Context) {
	if s.AccrualContainer != nil {
		_ = s.AccrualContainer.Terminate(ctx)
	}
}

// регистрирует товар в системе начислений
func (s *AccrualSuite) RegisterGood(ctx context.Context, match string, reward float64, rewardType string) error {
	url := s.AccrualURL + "/api/goods"
	payload := map[string]any{
		"match":       match,
		"reward":      reward,
		"reward_type": rewardType,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register good, status: %d", resp.StatusCode)
	}
	return nil
}

// регистрирует заказ в системе начислений
func (s *AccrualSuite) RegisterOrder(ctx context.Context, orderID string, goods []map[string]any) error {
	url := s.AccrualURL + "/api/orders"
	payload := map[string]any{
		"order": orderID,
		"goods": goods,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register order in accrual, status: %d", resp.StatusCode)
	}
	return nil
}

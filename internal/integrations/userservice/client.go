package userservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client клиент для работы с UserService
type Client struct {
	baseURL    string
	httpClient *http.Client
	log        Logger
}

// NewClient создает новый экземпляр клиента UserService
func NewClient(baseURL string, log Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}
}

// GetSuperUsers вызывает endpoint /internal/users/superusers для получения списка ID суперпользователей
func (c *Client) GetSuperUsers(ctx context.Context) ([]int64, error) {
	url := fmt.Sprintf("%s/internal/users/superusers", c.baseURL)

	// Создаём HTTP запрос
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrInternal, err)
	}

	// Выполняем запрос
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to execute request: %v", ErrInternal, err)
	}
	defer resp.Body.Close()

	// Обработка статус-кодов
	switch resp.StatusCode {
	case http.StatusOK:
		// Продолжаем обработку
	case http.StatusNotFound:
		return nil, ErrSuperUsersNotFound
	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: unexpected status code %d: %s", ErrInvalidResponse, resp.StatusCode, string(body))
	}

	// Парсим ответ
	var response SuperUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("%w: failed to decode response: %v", ErrInvalidResponse, err)
	}

	return response.SuperUserIDs, nil
}

// GetSuperUsersWithGracefulDegradation вызывает получение списка superusers с graceful degradation
// При недоступности UserService возвращает ErrServiceDegraded, что позволяет сервису продолжить работу без добавления superusers
func (c *Client) GetSuperUsersWithGracefulDegradation(ctx context.Context) ([]int64, error) {
	c.log.Info("Fetching superusers list from UserService")

	superUsers, err := c.GetSuperUsers(ctx)
	if err != nil {
		// Если это критичная бизнес-ошибка (не найдены superusers),
		// пробрасываем её дальше с оборачиванием
		if errors.Is(err, ErrSuperUsersNotFound) {
			c.log.Warn("SuperUsers not found in UserService")
			return nil, fmt.Errorf("userservice.client: GetSuperUsers failed - err: %w", err)
		}

		// Для всех остальных ошибок (недоступность сервиса, timeout, ошибки парсинга и т.д.)
		// применяем graceful degradation - возвращаем ErrServiceDegraded с контекстом
		c.log.Error("UserService unavailable, applying graceful degradation: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrServiceDegraded, err)
	}

	c.log.Info("Successfully fetched %d superusers from UserService", len(superUsers))
	return superUsers, nil
}

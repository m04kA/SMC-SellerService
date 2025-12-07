package userservice

import "errors"

var (
	// ErrSuperUsersNotFound возвращается, когда список суперпользователей не найден
	ErrSuperUsersNotFound = errors.New("superusers not found")

	// ErrInternal возвращается при внутренних ошибках клиента
	ErrInternal = errors.New("userservice client: internal error")

	// ErrInvalidResponse возвращается при некорректном ответе от сервиса
	ErrInvalidResponse = errors.New("userservice client: invalid response")

	// ErrServiceDegraded возвращается при применении graceful degradation
	// Указывает, что UserService недоступен и следует продолжить без списка superusers
	ErrServiceDegraded = errors.New("userservice unavailable: graceful degradation applied")
)

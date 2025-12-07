package userservice

// SuperUsersResponse ответ со списком ID суперпользователей
type SuperUsersResponse struct {
	SuperUserIDs []int64 `json:"super_user_ids"`
}

// ErrorResponse модель ошибки от UserService
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

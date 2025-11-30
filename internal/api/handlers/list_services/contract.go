package list_services

import (
	"context"

	"github.com/m04kA/SMC-SellerService/internal/service/services/models"
)

type ServiceService interface {
	ListByCompany(ctx context.Context, companyID int64, userID *int64) (*models.ServiceListResponse, error)
}

type Logger interface {
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

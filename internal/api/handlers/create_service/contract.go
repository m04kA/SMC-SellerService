package create_service

import (
	"context"

	"github.com/m04kA/SMC-SellerService/internal/service/services/models"
)

type ServiceService interface {
	Create(ctx context.Context, companyID int64, userID int64, userRole string, req *models.CreateServiceRequest) (*models.ServiceResponse, error)
}

type Logger interface {
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

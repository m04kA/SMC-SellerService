package update_service

import (
	"context"

	"github.com/m04kA/SMC-SellerService/internal/service/services/models"
)

type ServiceService interface {
	Update(ctx context.Context, companyID int64, serviceID int64, userID int64, userRole string, req *models.UpdateServiceRequest) (*models.ServiceResponse, error)
}

type Logger interface {
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

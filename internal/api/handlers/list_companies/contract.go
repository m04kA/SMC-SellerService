package list_companies

import (
	"context"

	"github.com/m04kA/SMC-SellerService/internal/service/companies/models"
)

type CompanyService interface {
	List(ctx context.Context, req *models.CompanyFilterRequest) (*models.CompanyListResponse, error)
}

type Logger interface {
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

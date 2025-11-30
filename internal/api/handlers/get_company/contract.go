package get_company

import (
	"context"

	"github.com/m04kA/SMC-SellerService/internal/service/companies/models"
)

type CompanyService interface {
	GetByID(ctx context.Context, id int64) (*models.CompanyResponse, error)
}

type Logger interface {
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

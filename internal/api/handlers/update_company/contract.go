package update_company

import (
	"context"

	"github.com/m04kA/SMC-SellerService/internal/service/companies/models"
)

type CompanyService interface {
	Update(ctx context.Context, id int64, userID int64, userRole string, req *models.UpdateCompanyRequest) (*models.CompanyResponse, error)
}

type Logger interface {
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}

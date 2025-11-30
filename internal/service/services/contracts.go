package services

import (
	"context"

	"github.com/m04kA/SMC-SellerService/internal/domain"
	"github.com/m04kA/SMC-SellerService/internal/integrations/priceservice"
)

// ServiceRepository интерфейс репозитория услуг
type ServiceRepository interface {
	Create(ctx context.Context, companyID int64, input domain.CreateServiceInput) (*domain.Service, error)
	GetByID(ctx context.Context, companyID int64, serviceID int64) (*domain.Service, error)
	ListByCompany(ctx context.Context, companyID int64) ([]domain.Service, error)
	Update(ctx context.Context, companyID int64, serviceID int64, input domain.UpdateServiceInput) (*domain.Service, error)
	Delete(ctx context.Context, companyID int64, serviceID int64) error
}

// CompanyRepository интерфейс для проверки прав доступа к компании
type CompanyRepository interface {
	IsManager(ctx context.Context, companyID int64, userID int64) (bool, error)
	GetByID(ctx context.Context, id int64) (*domain.Company, error)
}

// PriceServiceClient интерфейс для интеграции с PriceService
type PriceServiceClient interface {
	CalculatePricesWithGracefulDegradation(ctx context.Context, req *priceservice.CalculatePricesRequest) (*priceservice.CalculatePricesResponse, error)
}

package companies

import (
	"context"

	"github.com/m04kA/SMC-SellerService/internal/domain"
)

// CompanyRepository интерфейс репозитория компаний
type CompanyRepository interface {
	Create(ctx context.Context, input domain.CreateCompanyInput) (*domain.Company, error)
	GetByID(ctx context.Context, id int64) (*domain.Company, error)
	List(ctx context.Context, filter domain.CompanyFilter) ([]domain.Company, *domain.PaginationResult, error)
	Update(ctx context.Context, id int64, input domain.UpdateCompanyInput) (*domain.Company, error)
	Delete(ctx context.Context, id int64) error
	IsManager(ctx context.Context, companyID int64, userID int64) (bool, error)
}

// UserServiceClient интерфейс для работы с UserService
type UserServiceClient interface {
	GetSuperUsersWithGracefulDegradation(ctx context.Context) ([]int64, error)
}

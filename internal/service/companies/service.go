package companies

import (
	"context"
	"errors"
	"fmt"

	"github.com/m04kA/SMC-SellerService/internal/service"
	"github.com/m04kA/SMC-SellerService/internal/service/companies/models"
	companyRepo "github.com/m04kA/SMC-SellerService/internal/infra/storage/company"
)

type Service struct {
	companyRepo CompanyRepository
}

func NewService(companyRepo CompanyRepository) *Service {
	return &Service{
		companyRepo: companyRepo,
	}
}

// Create создает новую компанию
func (s *Service) Create(ctx context.Context, userID int64, userRole string, req *models.CreateCompanyRequest) (*models.CompanyResponse, error) {
	// Только superuser может создавать компании
	if userRole != service.RoleSuperuser {
		return nil, ErrOnlySuperuser
	}

	input := req.ToDomainCreateInput()
	company, err := s.companyRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%w: Create - repository error: %v", ErrInternal, err)
	}

	return models.FromDomainCompany(company), nil
}

// GetByID получает компанию по ID
func (s *Service) GetByID(ctx context.Context, id int64) (*models.CompanyResponse, error) {
	company, err := s.companyRepo.GetByID(ctx, id)
	if err != nil {
		// Проверяем, является ли ошибка ErrCompanyNotFound из репозитория
		if errors.Is(err, companyRepo.ErrCompanyNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, fmt.Errorf("%w: GetByID - repository error: %v", ErrInternal, err)
	}

	return models.FromDomainCompany(company), nil
}

// List получает список компаний с фильтрацией
func (s *Service) List(ctx context.Context, req *models.CompanyFilterRequest) (*models.CompanyListResponse, error) {
	filter := req.ToDomainFilter()
	companies, pagination, err := s.companyRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%w: List - repository error: %v", ErrInternal, err)
	}

	return models.FromDomainCompanyList(companies, pagination), nil
}

// Update обновляет компанию
func (s *Service) Update(ctx context.Context, id int64, userID int64, userRole string, req *models.UpdateCompanyRequest) (*models.CompanyResponse, error) {
	// Проверка прав доступа
	if err := s.checkAccess(ctx, id, userID, userRole); err != nil {
		return nil, err
	}

	input := req.ToDomainUpdateInput()
	company, err := s.companyRepo.Update(ctx, id, input)
	if err != nil {
		// Проверяем, является ли ошибка ErrCompanyNotFound из репозитория
		if errors.Is(err, companyRepo.ErrCompanyNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, fmt.Errorf("%w: Update - repository error: %v", ErrInternal, err)
	}

	return models.FromDomainCompany(company), nil
}

// Delete удаляет компанию
func (s *Service) Delete(ctx context.Context, id int64, userID int64, userRole string) error {
	// Только superuser может удалять компании
	if userRole != service.RoleSuperuser {
		return ErrOnlySuperuser
	}

	if err := s.companyRepo.Delete(ctx, id); err != nil {
		// Проверяем, является ли ошибка ErrCompanyNotFound из репозитория
		if errors.Is(err, companyRepo.ErrCompanyNotFound) {
			return ErrCompanyNotFound
		}
		return fmt.Errorf("%w: Delete - repository error: %v", ErrInternal, err)
	}

	return nil
}

// checkAccess проверяет права доступа пользователя к компании
func (s *Service) checkAccess(ctx context.Context, companyID int64, userID int64, userRole string) error {
	// Superuser имеет полный доступ
	if userRole == service.RoleSuperuser {
		return nil
	}

	// Обычный пользователь должен быть менеджером компании
	isManager, err := s.companyRepo.IsManager(ctx, companyID, userID)
	if err != nil {
		// Проверяем, является ли ошибка ErrCompanyNotFound из репозитория
		if errors.Is(err, companyRepo.ErrCompanyNotFound) {
			return ErrCompanyNotFound
		}
		return fmt.Errorf("%w: checkAccess - repository error: %v", ErrInternal, err)
	}

	if !isManager {
		return ErrAccessDenied
	}

	return nil
}

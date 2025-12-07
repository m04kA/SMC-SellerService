package companies

import (
	"context"
	"errors"
	"fmt"

	"github.com/m04kA/SMC-SellerService/internal/service"
	"github.com/m04kA/SMC-SellerService/internal/service/companies/models"
	companyRepo "github.com/m04kA/SMC-SellerService/internal/infra/storage/company"
	userServiceClient "github.com/m04kA/SMC-SellerService/internal/integrations/userservice"
)

type Service struct {
	companyRepo       CompanyRepository
	userServiceClient UserServiceClient
}

func NewService(companyRepo CompanyRepository, userServiceClient UserServiceClient) *Service {
	return &Service{
		companyRepo:       companyRepo,
		userServiceClient: userServiceClient,
	}
}

// Create создает новую компанию
func (s *Service) Create(ctx context.Context, userID int64, userRole string, req *models.CreateCompanyRequest) (*models.CompanyResponse, error) {
	// Только superuser может создавать компании
	if userRole != service.RoleSuperuser {
		return nil, ErrOnlySuperuser
	}

	input := req.ToDomainCreateInput()

	// Получаем список superusers и добавляем их в manager_ids
	superUsers, err := s.userServiceClient.GetSuperUsersWithGracefulDegradation(ctx)
	if err != nil {
		// Проверяем тип ошибки
		if !errors.Is(err, userServiceClient.ErrSuperUsersNotFound) && !errors.Is(err, userServiceClient.ErrServiceDegraded) {
			// Неизвестная критичная ошибка - прерываем создание
			return nil, fmt.Errorf("%w: failed to get superusers: %v", ErrInternal, err)
		}
	} else {
		// Объединяем переданные manager_ids с superusers (убираем дубликаты)
		input.ManagerIDs = s.mergeManagerIDs(input.ManagerIDs, superUsers)
	}

	company, err := s.companyRepo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%w: Create - repository error: %v", ErrInternal, err)
	}

	return models.FromDomainCompany(company), nil
}

func (s *Service) mergeManagerIDs(oldIds []int64, addIds []int64) []int64 {
	// Объединяем переданные manager_ids с superusers (убираем дубликаты)
	managerIDsMap := make(map[int64]bool)
	for _, id := range oldIds {
		managerIDsMap[id] = true
	}
	for _, id := range addIds {
		managerIDsMap[id] = true
	}

	// Преобразуем обратно в слайс
	mergedManagerIDs := make([]int64, 0, len(managerIDsMap))
	for id := range managerIDsMap {
		mergedManagerIDs = append(mergedManagerIDs, id)
	}
	return mergedManagerIDs
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

	// Получаем список superusers для проверки и добавления недостающих
	superUsers, err := s.userServiceClient.GetSuperUsersWithGracefulDegradation(ctx)
	if err != nil {
		// Проверяем тип ошибки
		if !errors.Is(err, userServiceClient.ErrSuperUsersNotFound) && !errors.Is(err, userServiceClient.ErrServiceDegraded) {
			// Неизвестная критичная ошибка - прерываем обновление
			return nil, fmt.Errorf("%w: failed to get superusers: %v", ErrInternal, err)
		}
	} else if len(input.ManagerIDs) > 0 {
		// Если manager_ids передавались явно - проверяем и добавляем недостающих superusers
		input.ManagerIDs = s.mergeManagerIDs(input.ManagerIDs, superUsers)
	} else if len(superUsers) > 0 {
		// Если manager_ids не передавались, но есть superusers - получаем текущие и добавляем недостающих
		currentCompany, err := s.companyRepo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, companyRepo.ErrCompanyNotFound) {
				return nil, ErrCompanyNotFound
			}
			return nil, fmt.Errorf("%w: Update - failed to get current company: %v", ErrInternal, err)
		}

		// Объединяем текущие manager_ids с superusers
		input.ManagerIDs = s.mergeManagerIDs(currentCompany.ManagerIDs, superUsers)
	}

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

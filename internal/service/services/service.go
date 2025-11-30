package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/m04kA/SMC-SellerService/internal/service"
	"github.com/m04kA/SMC-SellerService/internal/service/services/models"
	companyRepo "github.com/m04kA/SMC-SellerService/internal/infra/storage/company"
	serviceRepo "github.com/m04kA/SMC-SellerService/internal/infra/storage/service"
	"github.com/m04kA/SMC-SellerService/internal/integrations/priceservice"
)

type Service struct {
	serviceRepo ServiceRepository
	companyRepo CompanyRepository
	priceClient PriceServiceClient
}

func NewService(serviceRepo ServiceRepository, companyRepo CompanyRepository, priceClient PriceServiceClient) *Service {
	return &Service{
		serviceRepo: serviceRepo,
		companyRepo: companyRepo,
		priceClient: priceClient,
	}
}

// Create создает новую услугу для компании
func (s *Service) Create(ctx context.Context, companyID int64, userID int64, userRole string, req *models.CreateServiceRequest) (*models.ServiceResponse, error) {
	// Проверка прав доступа к компании
	if err := s.checkAccess(ctx, companyID, userID, userRole); err != nil {
		return nil, err
	}

	input := req.ToDomainCreateInput()
	service, err := s.serviceRepo.Create(ctx, companyID, input)
	if err != nil {
		return nil, fmt.Errorf("%w: Create - repository error: %v", ErrInternal, err)
	}

	return models.FromDomainService(service), nil
}

// GetByID получает услугу по ID с опциональным обогащением ценами
func (s *Service) GetByID(ctx context.Context, companyID int64, serviceID int64, userID *int64) (*models.ServiceResponse, error) {
	service, err := s.serviceRepo.GetByID(ctx, companyID, serviceID)
	if err != nil {
		if errors.Is(err, serviceRepo.ErrServiceNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, fmt.Errorf("%w: GetByID - repository error: %v", ErrInternal, err)
	}

	serviceDTO := models.FromDomainService(service)

	// Обогащаем ценами через PriceService
	servicePtrs := []*models.ServiceResponse{serviceDTO}
	// Graceful degradation: при ошибке PriceService возвращаем данные без цен
	s.enrichWithPrices(ctx, companyID, userID, servicePtrs)

	// Проверка на всякий случай (не должно произойти, но для безопасности)
	if len(servicePtrs) == 0 {
		return serviceDTO, nil
	}

	return servicePtrs[0], nil
}

// ListByCompany получает список услуг компании с опциональным обогащением ценами
func (s *Service) ListByCompany(ctx context.Context, companyID int64, userID *int64) (*models.ServiceListResponse, error) {
	services, err := s.serviceRepo.ListByCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("%w: ListByCompany - repository error: %v", ErrInternal, err)
	}

	listResponse := models.FromDomainServiceList(services)

	// Обогащаем ценами через PriceService
	if len(listResponse.Services) > 0 {
		// Создаём слайс указателей для обогащения
		servicePtrs := make([]*models.ServiceResponse, len(listResponse.Services))
		for i := range listResponse.Services {
			servicePtrs[i] = &listResponse.Services[i]
		}
		// Graceful degradation: при ошибке PriceService возвращаем данные без цен
		s.enrichWithPrices(ctx, companyID, userID, servicePtrs)

		// Обновляем список услуг из обогащённых указателей
		for i, svcPtr := range servicePtrs {
			if svcPtr != nil && i < len(listResponse.Services) {
				listResponse.Services[i] = *svcPtr
			}
		}
	}

	return listResponse, nil
}

// Update обновляет услугу
func (s *Service) Update(ctx context.Context, companyID int64, serviceID int64, userID int64, userRole string, req *models.UpdateServiceRequest) (*models.ServiceResponse, error) {
	// Проверка прав доступа к компании
	if err := s.checkAccess(ctx, companyID, userID, userRole); err != nil {
		return nil, err
	}

	input := req.ToDomainUpdateInput()
	service, err := s.serviceRepo.Update(ctx, companyID, serviceID, input)
	if err != nil {
		if errors.Is(err, serviceRepo.ErrServiceNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, fmt.Errorf("%w: Update - repository error: %v", ErrInternal, err)
	}

	return models.FromDomainService(service), nil
}

// Delete удаляет услугу
func (s *Service) Delete(ctx context.Context, companyID int64, serviceID int64, userID int64, userRole string) error {
	// Проверка прав доступа к компании
	if err := s.checkAccess(ctx, companyID, userID, userRole); err != nil {
		return err
	}

	if err := s.serviceRepo.Delete(ctx, companyID, serviceID); err != nil {
		if errors.Is(err, serviceRepo.ErrServiceNotFound) {
			return ErrServiceNotFound
		}
		return fmt.Errorf("%w: Delete - repository error: %v", ErrInternal, err)
	}

	return nil
}

// enrichWithPrices обогащает услуги ценами через PriceService
// При ошибке применяется graceful degradation - услуги возвращаются без цен
func (s *Service) enrichWithPrices(ctx context.Context, companyID int64, userID *int64, services []*models.ServiceResponse) {
	if len(services) == 0 {
		return
	}

	// Собираем ID услуг
	serviceIDs := make([]int64, len(services))
	for i, svc := range services {
		serviceIDs[i] = svc.ID
	}

	// Запрашиваем цены из PriceService
	pricesReq := &priceservice.CalculatePricesRequest{
		CompanyID:  companyID,
		UserID:     userID,
		ServiceIDs: serviceIDs,
	}

	pricesResp, err := s.priceClient.CalculatePricesWithGracefulDegradation(ctx, pricesReq)
	if err != nil {
		// Graceful degradation: если PriceService недоступен, просто не добавляем цены
		// Ошибка уже залогирована в клиенте PriceService
		return
	}

	// Создаём map для быстрого поиска цен по service_id
	priceMap := make(map[int64]priceservice.ServicePrice)
	for _, price := range pricesResp.Prices {
		priceMap[price.ServiceID] = price
	}

	// Обогащаем услуги ценами
	for _, svc := range services {
		if price, ok := priceMap[svc.ID]; ok {
			svc.EnrichWithPrice(
				price.Price,
				price.Currency,
				price.PricingType,
				price.VehicleClass,
				price.AppliedMultiplier,
			)
		}
	}
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

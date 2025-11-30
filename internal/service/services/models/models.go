package models

import (
	"time"

	"github.com/m04kA/SMC-SellerService/internal/domain"
)

// CreateServiceRequest запрос на создание услуги
type CreateServiceRequest struct {
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	AverageDuration *int    `json:"average_duration,omitempty"`
	AddressIDs      []int64 `json:"address_ids"`
}

// UpdateServiceRequest запрос на обновление услуги
type UpdateServiceRequest struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	AverageDuration *int    `json:"average_duration,omitempty"`
	AddressIDs      []int64 `json:"address_ids,omitempty"`
}

// ServiceResponse ответ с данными услуги
type ServiceResponse struct {
	ID              int64     `json:"id"`
	CompanyID       int64     `json:"company_id"`
	Name            string    `json:"name"`
	Description     *string   `json:"description,omitempty"`
	AverageDuration *int      `json:"average_duration,omitempty"`
	AddressIDs      []int64   `json:"address_ids"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	// Price fields (optional, populated when PriceService is available)
	Price             *float64 `json:"price,omitempty"`
	Currency          *string  `json:"currency,omitempty"`
	PricingType       *string  `json:"pricing_type,omitempty"`
	VehicleClass      *string  `json:"vehicle_class,omitempty"`
	AppliedMultiplier *float64 `json:"applied_multiplier,omitempty"`
}

// ServiceListResponse ответ со списком услуг
type ServiceListResponse struct {
	Services []ServiceResponse `json:"services"`
}

// ToDomainCreateInput конвертирует DTO в domain модель
func (r *CreateServiceRequest) ToDomainCreateInput() domain.CreateServiceInput {
	return domain.CreateServiceInput{
		Name:            r.Name,
		Description:     r.Description,
		AverageDuration: r.AverageDuration,
		AddressIDs:      r.AddressIDs,
	}
}

// ToDomainUpdateInput конвертирует DTO в domain модель
func (r *UpdateServiceRequest) ToDomainUpdateInput() domain.UpdateServiceInput {
	return domain.UpdateServiceInput{
		Name:            r.Name,
		Description:     r.Description,
		AverageDuration: r.AverageDuration,
		AddressIDs:      r.AddressIDs,
	}
}

// FromDomainService конвертирует domain модель в DTO
func FromDomainService(s *domain.Service) *ServiceResponse {
	return &ServiceResponse{
		ID:              s.ID,
		CompanyID:       s.CompanyID,
		Name:            s.Name,
		Description:     s.Description,
		AverageDuration: s.AverageDuration,
		AddressIDs:      s.AddressIDs,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
		// Price fields will be populated separately when needed
		Price:             nil,
		Currency:          nil,
		PricingType:       nil,
		VehicleClass:      nil,
		AppliedMultiplier: nil,
	}
}

// FromDomainServiceList конвертирует список domain моделей в DTO
func FromDomainServiceList(services []domain.Service) *ServiceListResponse {
	response := &ServiceListResponse{
		Services: make([]ServiceResponse, len(services)),
	}

	for i, s := range services {
		response.Services[i] = *FromDomainService(&s)
	}

	return response
}

// EnrichWithPrice обогащает ServiceResponse данными о цене
func (s *ServiceResponse) EnrichWithPrice(price *float64, currency *string, pricingType *string, vehicleClass *string, appliedMultiplier *float64) {
	s.Price = price
	s.Currency = currency
	s.PricingType = pricingType
	s.VehicleClass = vehicleClass
	s.AppliedMultiplier = appliedMultiplier
}

package models

import (
	"time"

	"github.com/m04kA/SMC-SellerService/internal/domain"
)

// CreateCompanyRequest запрос на создание компании
type CreateCompanyRequest struct {
	Name         string                `json:"name"`
	Logo         *string               `json:"logo,omitempty"`
	Description  *string               `json:"description,omitempty"`
	Tags         []string              `json:"tags"`
	Addresses    []AddressInput        `json:"addresses"`
	WorkingHours WorkingHoursInput     `json:"working_hours"`
	ManagerIDs   []int64               `json:"manager_ids"`
}

// UpdateCompanyRequest запрос на обновление компании
type UpdateCompanyRequest struct {
	Name         *string               `json:"name,omitempty"`
	Logo         *string               `json:"logo,omitempty"`
	Description  *string               `json:"description,omitempty"`
	Tags         []string              `json:"tags,omitempty"`
	Addresses    []AddressUpdateInput  `json:"addresses,omitempty"`
	WorkingHours *WorkingHoursInput    `json:"working_hours,omitempty"`
	ManagerIDs   []int64               `json:"manager_ids,omitempty"`
}

// AddressInput входные данные для адреса
type AddressInput struct {
	City        string      `json:"city"`
	Street      string      `json:"street"`
	Building    string      `json:"building"`
	Coordinates Coordinates `json:"coordinates"`
}

// AddressUpdateInput входные данные для обновления адреса
type AddressUpdateInput struct {
	ID          *int64      `json:"id,omitempty"`
	City        string      `json:"city"`
	Street      string      `json:"street"`
	Building    string      `json:"building"`
	Coordinates Coordinates `json:"coordinates"`
}

// Coordinates географические координаты
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// WorkingHoursInput входные данные для рабочих часов
type WorkingHoursInput struct {
	Monday    DaySchedule `json:"monday"`
	Tuesday   DaySchedule `json:"tuesday"`
	Wednesday DaySchedule `json:"wednesday"`
	Thursday  DaySchedule `json:"thursday"`
	Friday    DaySchedule `json:"friday"`
	Saturday  DaySchedule `json:"saturday"`
	Sunday    DaySchedule `json:"sunday"`
}

// DaySchedule расписание на день
type DaySchedule struct {
	IsOpen    bool    `json:"isOpen"`
	OpenTime  *string `json:"openTime,omitempty"`
	CloseTime *string `json:"closeTime,omitempty"`
}

// CompanyResponse ответ с данными компании
type CompanyResponse struct {
	ID           int64                 `json:"id"`
	Name         string                `json:"name"`
	Logo         *string               `json:"logo,omitempty"`
	Description  *string               `json:"description,omitempty"`
	Tags         []string              `json:"tags"`
	Addresses    []AddressResponse     `json:"addresses"`
	WorkingHours WorkingHoursResponse  `json:"working_hours"`
	ManagerIDs   []int64               `json:"manager_ids"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

// AddressResponse ответ с данными адреса
type AddressResponse struct {
	ID          int64       `json:"id"`
	City        string      `json:"city"`
	Street      string      `json:"street"`
	Building    string      `json:"building"`
	Coordinates Coordinates `json:"coordinates"`
}

// WorkingHoursResponse ответ с рабочими часами
type WorkingHoursResponse struct {
	Monday    DaySchedule `json:"monday"`
	Tuesday   DaySchedule `json:"tuesday"`
	Wednesday DaySchedule `json:"wednesday"`
	Thursday  DaySchedule `json:"thursday"`
	Friday    DaySchedule `json:"friday"`
	Saturday  DaySchedule `json:"saturday"`
	Sunday    DaySchedule `json:"sunday"`
}

// CompanyListResponse ответ со списком компаний
type CompanyListResponse struct {
	Companies  []CompanyResponse `json:"companies"`
	Pagination *PaginationResult `json:"pagination,omitempty"`
}

// PaginationResult результат пагинации
type PaginationResult struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
	TotalItems int `json:"total_items"`
}

// CompanyFilterRequest фильтр для списка компаний
type CompanyFilterRequest struct {
	Tags  []string `json:"tags,omitempty"`
	City  *string  `json:"city,omitempty"`
	Page  *int     `json:"page,omitempty"`
	Limit *int     `json:"limit,omitempty"`
}

// ToDomainCreateInput конвертирует DTO в domain модель
func (r *CreateCompanyRequest) ToDomainCreateInput() domain.CreateCompanyInput {
	addresses := make([]domain.AddressInput, len(r.Addresses))
	for i, addr := range r.Addresses {
		addresses[i] = domain.AddressInput{
			City:     addr.City,
			Street:   addr.Street,
			Building: addr.Building,
			Coordinates: domain.Coordinates{
				Latitude:  addr.Coordinates.Latitude,
				Longitude: addr.Coordinates.Longitude,
			},
		}
	}

	return domain.CreateCompanyInput{
		Name:        r.Name,
		Logo:        r.Logo,
		Description: r.Description,
		Tags:        r.Tags,
		Addresses:   addresses,
		WorkingHours: domain.WorkingHours{
			Monday:    toDomainDaySchedule(r.WorkingHours.Monday),
			Tuesday:   toDomainDaySchedule(r.WorkingHours.Tuesday),
			Wednesday: toDomainDaySchedule(r.WorkingHours.Wednesday),
			Thursday:  toDomainDaySchedule(r.WorkingHours.Thursday),
			Friday:    toDomainDaySchedule(r.WorkingHours.Friday),
			Saturday:  toDomainDaySchedule(r.WorkingHours.Saturday),
			Sunday:    toDomainDaySchedule(r.WorkingHours.Sunday),
		},
		ManagerIDs: r.ManagerIDs,
	}
}

// ToDomainUpdateInput конвертирует DTO в domain модель
func (r *UpdateCompanyRequest) ToDomainUpdateInput() domain.UpdateCompanyInput {
	var addresses []domain.AddressUpdateInput
	if r.Addresses != nil {
		addresses = make([]domain.AddressUpdateInput, len(r.Addresses))
		for i, addr := range r.Addresses {
			addresses[i] = domain.AddressUpdateInput{
				ID:       addr.ID,
				City:     addr.City,
				Street:   addr.Street,
				Building: addr.Building,
				Coordinates: domain.Coordinates{
					Latitude:  addr.Coordinates.Latitude,
					Longitude: addr.Coordinates.Longitude,
				},
			}
		}
	}

	var workingHours *domain.WorkingHours
	if r.WorkingHours != nil {
		workingHours = &domain.WorkingHours{
			Monday:    toDomainDaySchedule(r.WorkingHours.Monday),
			Tuesday:   toDomainDaySchedule(r.WorkingHours.Tuesday),
			Wednesday: toDomainDaySchedule(r.WorkingHours.Wednesday),
			Thursday:  toDomainDaySchedule(r.WorkingHours.Thursday),
			Friday:    toDomainDaySchedule(r.WorkingHours.Friday),
			Saturday:  toDomainDaySchedule(r.WorkingHours.Saturday),
			Sunday:    toDomainDaySchedule(r.WorkingHours.Sunday),
		}
	}

	return domain.UpdateCompanyInput{
		Name:         r.Name,
		Logo:         r.Logo,
		Description:  r.Description,
		Tags:         r.Tags,
		Addresses:    addresses,
		WorkingHours: workingHours,
		ManagerIDs:   r.ManagerIDs,
	}
}

// ToDomainFilter конвертирует DTO в domain модель
func (r *CompanyFilterRequest) ToDomainFilter() domain.CompanyFilter {
	return domain.CompanyFilter{
		Tags:  r.Tags,
		City:  r.City,
		Page:  r.Page,
		Limit: r.Limit,
	}
}

// FromDomainCompany конвертирует domain модель в DTO
func FromDomainCompany(c *domain.Company) *CompanyResponse {
	addresses := make([]AddressResponse, len(c.Addresses))
	for i, addr := range c.Addresses {
		addresses[i] = AddressResponse{
			ID:       addr.ID,
			City:     addr.City,
			Street:   addr.Street,
			Building: addr.Building,
			Coordinates: Coordinates{
				Latitude:  addr.Coordinates.Latitude,
				Longitude: addr.Coordinates.Longitude,
			},
		}
	}

	return &CompanyResponse{
		ID:          c.ID,
		Name:        c.Name,
		Logo:        c.Logo,
		Description: c.Description,
		Tags:        c.Tags,
		Addresses:   addresses,
		WorkingHours: WorkingHoursResponse{
			Monday:    fromDomainDaySchedule(c.WorkingHours.Monday),
			Tuesday:   fromDomainDaySchedule(c.WorkingHours.Tuesday),
			Wednesday: fromDomainDaySchedule(c.WorkingHours.Wednesday),
			Thursday:  fromDomainDaySchedule(c.WorkingHours.Thursday),
			Friday:    fromDomainDaySchedule(c.WorkingHours.Friday),
			Saturday:  fromDomainDaySchedule(c.WorkingHours.Saturday),
			Sunday:    fromDomainDaySchedule(c.WorkingHours.Sunday),
		},
		ManagerIDs: c.ManagerIDs,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

// FromDomainCompanyList конвертирует список domain моделей в DTO
func FromDomainCompanyList(companies []domain.Company, pagination *domain.PaginationResult) *CompanyListResponse {
	response := &CompanyListResponse{
		Companies: make([]CompanyResponse, len(companies)),
	}

	for i, c := range companies {
		response.Companies[i] = *FromDomainCompany(&c)
	}

	if pagination != nil {
		totalPages := (pagination.Total + pagination.Limit - 1) / pagination.Limit
		response.Pagination = &PaginationResult{
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			TotalPages: totalPages,
			TotalItems: pagination.Total,
		}
	}

	return response
}

func toDomainDaySchedule(ds DaySchedule) domain.DaySchedule {
	return domain.DaySchedule{
		IsOpen:    ds.IsOpen,
		OpenTime:  stringPtrToTimeString(ds.OpenTime),
		CloseTime: stringPtrToTimeString(ds.CloseTime),
	}
}

func fromDomainDaySchedule(ds domain.DaySchedule) DaySchedule {
	return DaySchedule{
		IsOpen:    ds.IsOpen,
		OpenTime:  timeStringToStringPtr(ds.OpenTime),
		CloseTime: timeStringToStringPtr(ds.CloseTime),
	}
}

func stringPtrToTimeString(s *string) *domain.TimeString {
	if s == nil {
		return nil
	}
	ts := domain.TimeString(*s)
	return &ts
}

func timeStringToStringPtr(ts *domain.TimeString) *string {
	if ts == nil {
		return nil
	}
	s := string(*ts)
	return &s
}

package get_service

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers"
	"github.com/m04kA/SMC-SellerService/internal/api/middleware"
	"github.com/m04kA/SMC-SellerService/internal/service/services"
)

const (
	msgInvalidCompanyID = "invalid company ID"
	msgInvalidServiceID = "invalid service ID"
	msgNotFound         = "service not found"
)

type Handler struct {
	service ServiceService
	logger  Logger
}

func NewHandler(service ServiceService, logger Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Handle GET /api/v1/companies/{company_id}/services/{service_id}
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	companyIDStr := vars["company_id"]
	serviceIDStr := vars["service_id"]

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("GET /companies/{company_id}/services/{service_id} - Invalid company ID: %v", err)
		handlers.RespondBadRequest(w, msgInvalidCompanyID)
		return
	}

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("GET /companies/{company_id}/services/{service_id} - Invalid service ID: %v", err)
		handlers.RespondBadRequest(w, msgInvalidServiceID)
		return
	}

	// Получаем опциональный userID из контекста (через OptionalAuth middleware)
	var userID *int64
	if ctxUserID, ok := middleware.GetUserID(r.Context()); ok && ctxUserID > 0 {
		userID = &ctxUserID
	}

	service, err := h.service.GetByID(r.Context(), companyID, serviceID, userID)
	if err != nil {
		if errors.Is(err, services.ErrServiceNotFound) {
			h.logger.Warn("GET /companies/{company_id}/services/{service_id} - Service not found: company_id=%d, service_id=%d", companyID, serviceID)
			handlers.RespondNotFound(w, msgNotFound)
			return
		}
		h.logger.Error("GET /companies/{company_id}/services/{service_id} - Failed to get service: company_id=%d, service_id=%d, error=%v", companyID, serviceID, err)
		handlers.RespondInternalError(w)
		return
	}

	if userID != nil {
		h.logger.Info("GET /companies/{company_id}/services/{service_id} - Service retrieved successfully: company_id=%d, service_id=%d, user_id=%d", companyID, serviceID, *userID)
	} else {
		h.logger.Info("GET /companies/{company_id}/services/{service_id} - Service retrieved successfully: company_id=%d, service_id=%d", companyID, serviceID)
	}
	handlers.RespondJSON(w, http.StatusOK, service)
}

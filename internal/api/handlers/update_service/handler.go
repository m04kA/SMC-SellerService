package update_service

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers"
	"github.com/m04kA/SMC-SellerService/internal/api/middleware"
	"github.com/m04kA/SMC-SellerService/internal/service/services"
	"github.com/m04kA/SMC-SellerService/internal/service/services/models"
)

const (
	msgInvalidRequestBody = "invalid request body"
	msgInvalidCompanyID   = "invalid company ID"
	msgInvalidServiceID   = "invalid service ID"
	msgForbidden          = "access denied"
	msgNotFound           = "service not found"
	msgCompanyNotFound    = "company not found"
	msgMissingUserID      = "missing user ID"
	msgMissingUserRole    = "missing user role"
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

// Handle PUT /api/v1/companies/{company_id}/services/{service_id}
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		handlers.RespondUnauthorized(w, msgMissingUserID)
		return
	}

	userRole, ok := middleware.GetUserRole(r.Context())
	if !ok {
		handlers.RespondUnauthorized(w, msgMissingUserRole)
		return
	}

	vars := mux.Vars(r)
	companyIDStr := vars["company_id"]
	serviceIDStr := vars["service_id"]

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("PUT /companies/{company_id}/services/{service_id} - Invalid company ID: %v", err)
		handlers.RespondBadRequest(w, msgInvalidCompanyID)
		return
	}

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("PUT /companies/{company_id}/services/{service_id} - Invalid service ID: %v", err)
		handlers.RespondBadRequest(w, msgInvalidServiceID)
		return
	}

	var req models.UpdateServiceRequest
	if err := handlers.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("PUT /companies/{company_id}/services/{service_id} - Invalid request body: %v", err)
		handlers.RespondBadRequest(w, msgInvalidRequestBody)
		return
	}

	service, err := h.service.Update(r.Context(), companyID, serviceID, userID, userRole, &req)
	if err != nil {
		if errors.Is(err, services.ErrServiceNotFound) {
			h.logger.Warn("PUT /companies/{company_id}/services/{service_id} - Service not found: company_id=%d, service_id=%d", companyID, serviceID)
			handlers.RespondNotFound(w, msgNotFound)
			return
		}
		if errors.Is(err, services.ErrCompanyNotFound) {
			h.logger.Warn("PUT /companies/{company_id}/services/{service_id} - Company not found: company_id=%d", companyID)
			handlers.RespondNotFound(w, msgCompanyNotFound)
			return
		}
		if errors.Is(err, services.ErrAccessDenied) {
			h.logger.Warn("PUT /companies/{company_id}/services/{service_id} - Access denied: company_id=%d, service_id=%d, user_id=%d", companyID, serviceID, userID)
			handlers.RespondForbidden(w, msgForbidden)
			return
		}
		h.logger.Error("PUT /companies/{company_id}/services/{service_id} - Failed to update service: company_id=%d, service_id=%d, user_id=%d, error=%v", companyID, serviceID, userID, err)
		handlers.RespondInternalError(w)
		return
	}

	h.logger.Info("PUT /companies/{company_id}/services/{service_id} - Service updated successfully: company_id=%d, service_id=%d, user_id=%d", companyID, serviceID, userID)
	handlers.RespondJSON(w, http.StatusOK, service)
}

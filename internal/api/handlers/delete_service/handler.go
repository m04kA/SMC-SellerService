package delete_service

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
	msgForbidden        = "access denied"
	msgNotFound         = "service not found"
	msgCompanyNotFound  = "company not found"
	msgMissingUserID    = "missing user ID"
	msgMissingUserRole  = "missing user role"
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

// Handle DELETE /api/v1/companies/{company_id}/services/{service_id}
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
		h.logger.Warn("DELETE /companies/{company_id}/services/{service_id} - Invalid company ID: %v", err)
		handlers.RespondBadRequest(w, msgInvalidCompanyID)
		return
	}

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("DELETE /companies/{company_id}/services/{service_id} - Invalid service ID: %v", err)
		handlers.RespondBadRequest(w, msgInvalidServiceID)
		return
	}

	err = h.service.Delete(r.Context(), companyID, serviceID, userID, userRole)
	if err != nil {
		if errors.Is(err, services.ErrServiceNotFound) {
			h.logger.Warn("DELETE /companies/{company_id}/services/{service_id} - Service not found: company_id=%d, service_id=%d", companyID, serviceID)
			handlers.RespondNotFound(w, msgNotFound)
			return
		}
		if errors.Is(err, services.ErrCompanyNotFound) {
			h.logger.Warn("DELETE /companies/{company_id}/services/{service_id} - Company not found: company_id=%d", companyID)
			handlers.RespondNotFound(w, msgCompanyNotFound)
			return
		}
		if errors.Is(err, services.ErrAccessDenied) {
			h.logger.Warn("DELETE /companies/{company_id}/services/{service_id} - Access denied: company_id=%d, service_id=%d, user_id=%d", companyID, serviceID, userID)
			handlers.RespondForbidden(w, msgForbidden)
			return
		}
		h.logger.Error("DELETE /companies/{company_id}/services/{service_id} - Failed to delete service: company_id=%d, service_id=%d, user_id=%d, error=%v", companyID, serviceID, userID, err)
		handlers.RespondInternalError(w)
		return
	}

	h.logger.Info("DELETE /companies/{company_id}/services/{service_id} - Service deleted successfully: company_id=%d, service_id=%d, user_id=%d", companyID, serviceID, userID)
	w.WriteHeader(http.StatusNoContent)
}

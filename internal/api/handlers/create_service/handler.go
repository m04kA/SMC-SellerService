package create_service

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
	msgForbidden          = "access denied"
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

// Handle POST /api/v1/companies/{company_id}/services
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

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("POST /companies/{company_id}/services - Invalid company ID: %v", err)
		handlers.RespondBadRequest(w, msgInvalidCompanyID)
		return
	}

	var req models.CreateServiceRequest
	if err := handlers.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("POST /companies/{company_id}/services - Invalid request body: %v", err)
		handlers.RespondBadRequest(w, msgInvalidRequestBody)
		return
	}

	service, err := h.service.Create(r.Context(), companyID, userID, userRole, &req)
	if err != nil {
		if errors.Is(err, services.ErrCompanyNotFound) {
			h.logger.Warn("POST /companies/{company_id}/services - Company not found: company_id=%d", companyID)
			handlers.RespondNotFound(w, msgCompanyNotFound)
			return
		}
		if errors.Is(err, services.ErrAccessDenied) {
			h.logger.Warn("POST /companies/{company_id}/services - Access denied: company_id=%d, user_id=%d", companyID, userID)
			handlers.RespondForbidden(w, msgForbidden)
			return
		}
		h.logger.Error("POST /companies/{company_id}/services - Failed to create service: company_id=%d, user_id=%d, error=%v", companyID, userID, err)
		handlers.RespondInternalError(w)
		return
	}

	h.logger.Info("POST /companies/{company_id}/services - Service created successfully: service_id=%d, company_id=%d, user_id=%d", service.ID, companyID, userID)
	handlers.RespondJSON(w, http.StatusCreated, service)
}

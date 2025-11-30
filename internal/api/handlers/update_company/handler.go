package update_company

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers"
	"github.com/m04kA/SMC-SellerService/internal/api/middleware"
	"github.com/m04kA/SMC-SellerService/internal/service/companies"
	"github.com/m04kA/SMC-SellerService/internal/service/companies/models"
)

const (
	msgInvalidRequestBody = "invalid request body"
	msgInvalidCompanyID   = "invalid company ID"
	msgForbidden          = "access denied"
	msgNotFound           = "company not found"
	msgMissingUserID      = "missing user ID"
	msgMissingUserRole    = "missing user role"
)

type Handler struct {
	service CompanyService
	logger  Logger
}

func NewHandler(service CompanyService, logger Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Handle PUT /api/v1/companies/{id}
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
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("PUT /companies/{id} - Invalid company ID: %v", err)
		handlers.RespondBadRequest(w, msgInvalidCompanyID)
		return
	}

	var req models.UpdateCompanyRequest
	if err := handlers.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("PUT /companies/{id} - Invalid request body: %v", err)
		handlers.RespondBadRequest(w, msgInvalidRequestBody)
		return
	}

	company, err := h.service.Update(r.Context(), id, userID, userRole, &req)
	if err != nil {
		if errors.Is(err, companies.ErrCompanyNotFound) {
			h.logger.Warn("PUT /companies/{id} - Company not found: company_id=%d", id)
			handlers.RespondNotFound(w, msgNotFound)
			return
		}
		if errors.Is(err, companies.ErrAccessDenied) {
			h.logger.Warn("PUT /companies/{id} - Access denied: company_id=%d, user_id=%d", id, userID)
			handlers.RespondForbidden(w, msgForbidden)
			return
		}
		h.logger.Error("PUT /companies/{id} - Failed to update company: company_id=%d, user_id=%d, error=%v", id, userID, err)
		handlers.RespondInternalError(w)
		return
	}

	h.logger.Info("PUT /companies/{id} - Company updated successfully: company_id=%d, user_id=%d", id, userID)
	handlers.RespondJSON(w, http.StatusOK, company)
}

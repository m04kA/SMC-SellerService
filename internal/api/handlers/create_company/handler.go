package create_company

import (
	"errors"
	"net/http"

	"github.com/m04kA/SMC-SellerService/internal/api/handlers"
	"github.com/m04kA/SMC-SellerService/internal/api/middleware"
	"github.com/m04kA/SMC-SellerService/internal/service/companies"
	"github.com/m04kA/SMC-SellerService/internal/service/companies/models"
)

const (
	msgInvalidRequestBody = "invalid request body"
	msgForbidden          = "access denied"
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

// Handle POST /api/v1/companies
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

	var req models.CreateCompanyRequest
	if err := handlers.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("POST /companies - Invalid request body: %v", err)
		handlers.RespondBadRequest(w, msgInvalidRequestBody)
		return
	}

	company, err := h.service.Create(r.Context(), userID, userRole, &req)
	if err != nil {
		if errors.Is(err, companies.ErrOnlySuperuser) {
			h.logger.Warn("POST /companies - Access denied: user_id=%d, role=%s", userID, userRole)
			handlers.RespondForbidden(w, msgForbidden)
			return
		}
		if errors.Is(err, companies.ErrAccessDenied) {
			h.logger.Warn("POST /companies - Access denied: user_id=%d", userID)
			handlers.RespondForbidden(w, msgForbidden)
			return
		}
		h.logger.Error("POST /companies - Failed to create company: user_id=%d, error=%v", userID, err)
		handlers.RespondInternalError(w)
		return
	}

	h.logger.Info("POST /companies - Company created successfully: company_id=%d, user_id=%d", company.ID, userID)
	handlers.RespondJSON(w, http.StatusCreated, company)
}

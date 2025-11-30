package delete_company

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers"
	"github.com/m04kA/SMC-SellerService/internal/api/middleware"
	"github.com/m04kA/SMC-SellerService/internal/service/companies"
)

const (
	msgInvalidCompanyID = "invalid company ID"
	msgForbidden        = "access denied"
	msgNotFound         = "company not found"
	msgMissingUserID    = "missing user ID"
	msgMissingUserRole  = "missing user role"
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

// Handle DELETE /api/v1/companies/{id}
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
		h.logger.Warn("DELETE /companies/{id} - Invalid company ID: %v", err)
		handlers.RespondBadRequest(w, msgInvalidCompanyID)
		return
	}

	err = h.service.Delete(r.Context(), id, userID, userRole)
	if err != nil {
		if errors.Is(err, companies.ErrCompanyNotFound) {
			h.logger.Warn("DELETE /companies/{id} - Company not found: company_id=%d", id)
			handlers.RespondNotFound(w, msgNotFound)
			return
		}
		if errors.Is(err, companies.ErrOnlySuperuser) {
			h.logger.Warn("DELETE /companies/{id} - Access denied: company_id=%d, user_id=%d, role=%s", id, userID, userRole)
			handlers.RespondForbidden(w, msgForbidden)
			return
		}
		h.logger.Error("DELETE /companies/{id} - Failed to delete company: company_id=%d, user_id=%d, error=%v", id, userID, err)
		handlers.RespondInternalError(w)
		return
	}

	h.logger.Info("DELETE /companies/{id} - Company deleted successfully: company_id=%d, user_id=%d", id, userID)
	w.WriteHeader(http.StatusNoContent)
}

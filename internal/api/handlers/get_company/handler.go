package get_company

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers"
	"github.com/m04kA/SMC-SellerService/internal/service/companies"
)

const (
	msgInvalidCompanyID = "invalid company ID"
	msgNotFound         = "company not found"
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

// Handle GET /api/v1/companies/{id}
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("GET /companies/{id} - Invalid company ID: %v", err)
		handlers.RespondBadRequest(w, msgInvalidCompanyID)
		return
	}

	company, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, companies.ErrCompanyNotFound) {
			h.logger.Warn("GET /companies/{id} - Company not found: company_id=%d", id)
			handlers.RespondNotFound(w, msgNotFound)
			return
		}
		h.logger.Error("GET /companies/{id} - Failed to get company: company_id=%d, error=%v", id, err)
		handlers.RespondInternalError(w)
		return
	}

	h.logger.Info("GET /companies/{id} - Company retrieved successfully: company_id=%d", id)
	handlers.RespondJSON(w, http.StatusOK, company)
}

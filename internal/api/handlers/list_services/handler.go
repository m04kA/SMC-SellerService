package list_services

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/m04kA/SMC-SellerService/internal/api/handlers"
	"github.com/m04kA/SMC-SellerService/internal/api/middleware"
)

const (
	msgInvalidCompanyID = "invalid company ID"
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

// Handle GET /api/v1/companies/{company_id}/services
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	companyIDStr := vars["company_id"]

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("GET /companies/{company_id}/services - Invalid company ID: %v", err)
		handlers.RespondBadRequest(w, msgInvalidCompanyID)
		return
	}

	// Получаем опциональный userID из контекста (через OptionalAuth middleware)
	var userID *int64
	if ctxUserID, ok := middleware.GetUserID(r.Context()); ok && ctxUserID > 0 {
		userID = &ctxUserID
	}

	response, err := h.service.ListByCompany(r.Context(), companyID, userID)
	if err != nil {
		h.logger.Error("GET /companies/{company_id}/services - Failed to list services: company_id=%d, error=%v", companyID, err)
		handlers.RespondInternalError(w)
		return
	}

	if userID != nil {
		h.logger.Info("GET /companies/{company_id}/services - Services listed successfully: company_id=%d, user_id=%d, count=%d", companyID, *userID, len(response.Services))
	} else {
		h.logger.Info("GET /companies/{company_id}/services - Services listed successfully: company_id=%d, count=%d", companyID, len(response.Services))
	}
	handlers.RespondJSON(w, http.StatusOK, response)
}

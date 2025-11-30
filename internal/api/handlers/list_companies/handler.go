package list_companies

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/m04kA/SMC-SellerService/internal/api/handlers"
	"github.com/m04kA/SMC-SellerService/internal/service/companies/models"
)

const (
	msgInvalidPageParam  = "invalid page parameter"
	msgInvalidLimitParam = "invalid limit parameter"
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

// Handle GET /api/v1/companies
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Парсим фильтры
	var req models.CompanyFilterRequest

	// Парсим теги (опционально)
	if tagsStr := query.Get("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		req.Tags = tags
	}

	// Парсим город (опционально)
	if city := query.Get("city"); city != "" {
		req.City = &city
	}

	// Парсим пагинацию (опционально)
	if pageStr := query.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			h.logger.Warn("GET /companies - Invalid page parameter: %v", err)
			handlers.RespondBadRequest(w, msgInvalidPageParam)
			return
		}
		req.Page = &page
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			h.logger.Warn("GET /companies - Invalid limit parameter: %v", err)
			handlers.RespondBadRequest(w, msgInvalidLimitParam)
			return
		}
		req.Limit = &limit
	}

	response, err := h.service.List(r.Context(), &req)
	if err != nil {
		h.logger.Error("GET /companies - Failed to list companies: error=%v", err)
		handlers.RespondInternalError(w)
		return
	}

	h.logger.Info("GET /companies - Companies listed successfully: count=%d", len(response.Companies))
	handlers.RespondJSON(w, http.StatusOK, response)
}

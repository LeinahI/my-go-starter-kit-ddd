package http

import (
	"encoding/json"
	"errors"
	"net/http"

	appproduct "github.com/yourorg/ws/internal/application/product"
	"github.com/yourorg/ws/internal/domain/product"
	"github.com/yourorg/ws/internal/transport/http/response"
)

type ProductHandler struct {
	svc *appproduct.Service
}

func NewProductHandler(svc *appproduct.Service) *ProductHandler {
	return &ProductHandler{svc: svc}
}

type productResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	Price       string `json:"price"`
	Stock       int    `json:"stock"`
}

type createProductRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Stock       int    `json:"stock"`
}

// List godoc
//
//	@Summary		List products
//	@Description	Returns a paginated list of products. Optional search via query parameter.
//	@Tags			products
//	@Produce		json
//	@Param			q	query		string	false	"Search by name or slug"
//	@Success		200	{object}	SwaggerProductListEnvelope
//	@Failure		500	{object}	SwaggerErrorEnvelope
//	@Router			/api/v1/products [get]
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context(), product.ListFilter{
		Query: r.URL.Query().Get("q"),
		Limit: 20,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list products")
		return
	}

	out := make([]productResponse, 0, len(items))
	for _, p := range items {
		out = append(out, toProductResponse(p))
	}
	response.JSON(w, http.StatusOK, out)
}

// Get godoc
//
//	@Summary		Get product by ID
//	@Tags			products
//	@Produce		json
//	@Param			id	path		string	true	"Product UUID"
//	@Success		200	{object}	SwaggerProductEnvelope
//	@Failure		404	{object}	SwaggerErrorEnvelope
//	@Failure		500	{object}	SwaggerErrorEnvelope
//	@Router			/api/v1/products/{id} [get]
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeProductError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, toProductResponse(p))
}

// Create godoc
//
//	@Summary		Create product
//	@Tags			products
//	@Accept			json
//	@Produce		json
//	@Param			body	body		SwaggerCreateProductRequest	true	"Product payload"
//	@Success		201	{object}	SwaggerProductEnvelope
//	@Failure		400	{object}	SwaggerErrorEnvelope
//	@Failure		500	{object}	SwaggerErrorEnvelope
//	@Router			/api/v1/products [post]
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	p, err := h.svc.Create(r.Context(), appproduct.CreateInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	})
	if err != nil {
		writeProductError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, toProductResponse(p))
}

func toProductResponse(p *product.Product) productResponse {
	return productResponse{
		ID:          p.ID(),
		Name:        p.Name(),
		Slug:        p.Slug(),
		Description: p.Description(),
		Price:       p.Price().Amount(),
		Stock:       p.Stock(),
	}
}

func writeProductError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, product.ErrNotFound):
		response.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, product.ErrDuplicateSlug),
		errors.Is(err, product.ErrInvalidName),
		errors.Is(err, product.ErrInvalidSlug),
		errors.Is(err, product.ErrInvalidStock):
		response.Error(w, http.StatusBadRequest, err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

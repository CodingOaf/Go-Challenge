package catalog

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
	"gorm.io/gorm"
)

type ProductsRepository interface {
	GetProducts(offset, limit int, category string, priceLessThan *float64) ([]models.Product, int64, error)
	GetProductByCode(code string) (*models.Product, error)
}

type CategoriesRepository interface {
	GetAllCategories() ([]models.Category, error)
	CreateCategory(category *models.Category) error
}

type CatalogHandler struct {
	products   ProductsRepository
	categories CategoriesRepository
}

func NewCatalogHandler(products ProductsRepository, categories CategoriesRepository) *CatalogHandler {
	return &CatalogHandler{products: products, categories: categories}
}

type CategoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProductResponse struct {
	Code     string           `json:"code"`
	Price    float64          `json:"price"`
	Category CategoryResponse `json:"category"`
}

type CatalogResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int64             `json:"total"`
	Offset   int               `json:"offset"`
	Limit    int               `json:"limit"`
}

type VariantResponse struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

type ProductDetailsResponse struct {
	Code     string            `json:"code"`
	Price    float64           `json:"price"`
	Category CategoryResponse  `json:"category"`
	Variants []VariantResponse `json:"variants"`
}

type CategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

type CategoryCreateRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (h *CatalogHandler) HandleGetCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query()
	offset := 0
	if q := query.Get("offset"); q != "" {
		parsed, err := strconv.Atoi(q)
		if err != nil || parsed < 0 {
			api.ErrorResponse(w, http.StatusBadRequest, "offset must be a non-negative integer")
			return
		}
		offset = parsed
	}

	limit := 10
	if q := query.Get("limit"); q != "" {
		parsed, err := strconv.Atoi(q)
		if err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, "limit must be an integer")
			return
		}
		if parsed < 1 || parsed > 100 {
			api.ErrorResponse(w, http.StatusBadRequest, "limit must be between 1 and 100")
			return
		}
		limit = parsed
	}

	priceLessThanStr := query.Get("price_less_than")
	if priceLessThanStr == "" {
		priceLessThanStr = query.Get("priceLessThan")
	}

	var priceLessThan *float64
	if priceLessThanStr != "" {
		parsed, err := strconv.ParseFloat(priceLessThanStr, 64)
		if err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, "price_less_than must be a valid decimal")
			return
		}
		priceLessThan = &parsed
	}

	products, total, err := h.products.GetProducts(offset, limit, query.Get("category"), priceLessThan)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch products")
		return
	}

	responseProducts := make([]ProductResponse, len(products))
	for i, p := range products {
		responseProducts[i] = ProductResponse{
			Code:  p.Code,
			Price: p.Price,
			Category: CategoryResponse{
				Code: p.Category.Code,
				Name: p.Category.Name,
			},
		}
	}

	api.OKResponse(w, CatalogResponse{
		Products: responseProducts,
		Total:    total,
		Offset:   offset,
		Limit:    limit,
	})
}

func (h *CatalogHandler) HandleGetProductDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	code := strings.TrimPrefix(r.URL.Path, "/catalog/")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "product code is required")
		return
	}
	code = path.Clean(code)
	if strings.Contains(code, "/") {
		code = strings.TrimSuffix(code, "/")
	}

	product, err := h.products.GetProductByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			api.ErrorResponse(w, http.StatusNotFound, "product not found")
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch product")
		return
	}

	variants := make([]VariantResponse, len(product.Variants))
	for i, variant := range product.Variants {
		price := product.Price
		if variant.Price != nil {
			price = *variant.Price
		}
		variants[i] = VariantResponse{
			Name:  variant.Name,
			SKU:   variant.SKU,
			Price: price,
		}
	}

	api.OKResponse(w, ProductDetailsResponse{
		Code:  product.Code,
		Price: product.Price,
		Category: CategoryResponse{
			Code: product.Category.Code,
			Name: product.Category.Name,
		},
		Variants: variants,
	})
}

func (h *CatalogHandler) HandleCategories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		categories, err := h.categories.GetAllCategories()
		if err != nil {
			api.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch categories")
			return
		}

		responseCategories := make([]CategoryResponse, len(categories))
		for i, category := range categories {
			responseCategories[i] = CategoryResponse{Code: category.Code, Name: category.Name}
		}

		api.OKResponse(w, CategoriesResponse{Categories: responseCategories})

	case http.MethodPost:
		var request CategoryCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
			return
		}

		request.Code = strings.TrimSpace(request.Code)
		request.Name = strings.TrimSpace(request.Name)
		if request.Code == "" || request.Name == "" {
			api.ErrorResponse(w, http.StatusBadRequest, "code and name are required")
			return
		}

		category := models.Category{Code: request.Code, Name: request.Name}
		if err := h.categories.CreateCategory(&category); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				api.ErrorResponse(w, http.StatusBadRequest, "category code already exists")
				return
			}
			api.ErrorResponse(w, http.StatusInternalServerError, "failed to create category")
			return
		}

		api.OKResponse(w, CategoryResponse{Code: category.Code, Name: category.Name})
	default:
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

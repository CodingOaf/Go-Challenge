package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type fakeProductsRepository struct {
	products  []models.Product
	total     int64
	product   *models.Product
	listErr   error
	detailErr error
}

func (f *fakeProductsRepository) GetProducts(offset, limit int, category string, priceLessThan *float64) ([]models.Product, int64, error) {
	return f.products, f.total, f.listErr
}

func (f *fakeProductsRepository) GetProductByCode(code string) (*models.Product, error) {
	return f.product, f.detailErr
}

type fakeCategoriesRepository struct {
	categories []models.Category
	createErr  error
}

func (f *fakeCategoriesRepository) GetAllCategories() ([]models.Category, error) {
	return f.categories, nil
}

func (f *fakeCategoriesRepository) CreateCategory(category *models.Category) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.categories = append(f.categories, *category)
	return nil
}

func floatPtr(value float64) *float64 {
	return &value
}

func TestHandleGetCatalog(t *testing.T) {
	fakeRepo := &fakeProductsRepository{
		products: []models.Product{
			{
				Code:     "PROD001",
				Price:    10.99,
				Category: models.Category{Code: "clothing", Name: "Clothing"},
			},
		},
		total: 1,
	}
	handler := NewCatalogHandler(fakeRepo, &fakeCategoriesRepository{})

	req := httptest.NewRequest(http.MethodGet, "/catalog?offset=0&limit=10&category=clothing&price_less_than=20", nil)
	recorder := httptest.NewRecorder()

	handler.HandleGetCatalog(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response struct {
		Products []struct {
			Code     string  `json:"code"`
			Price    float64 `json:"price"`
			Category struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"category"`
		} `json:"products"`
		Total  int64 `json:"total"`
		Offset int   `json:"offset"`
		Limit  int   `json:"limit"`
	}

	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response.Products, 1)
	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, 0, response.Offset)
	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, "PROD001", response.Products[0].Code)
	assert.Equal(t, "clothing", response.Products[0].Category.Code)
}

func TestHandleGetCatalog_InvalidLimit(t *testing.T) {
	handler := NewCatalogHandler(&fakeProductsRepository{}, &fakeCategoriesRepository{})

	req := httptest.NewRequest(http.MethodGet, "/catalog?limit=0", nil)
	recorder := httptest.NewRecorder()

	handler.HandleGetCatalog(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestHandleGetProductDetails(t *testing.T) {
	fakeRepo := &fakeProductsRepository{
		product: &models.Product{
			Code:     "PROD001",
			Price:    10.99,
			Category: models.Category{Code: "clothing", Name: "Clothing"},
			Variants: []models.Variant{
				{Name: "Variant A", SKU: "SKU001A", Price: floatPtr(11.99)},
				{Name: "Variant B", SKU: "SKU001B", Price: nil},
			},
		},
	}
	handler := NewCatalogHandler(fakeRepo, &fakeCategoriesRepository{})

	req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
	recorder := httptest.NewRecorder()

	handler.HandleGetProductDetails(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response struct {
		Code     string  `json:"code"`
		Price    float64 `json:"price"`
		Category struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"category"`
		Variants []struct {
			Name  string  `json:"name"`
			SKU   string  `json:"sku"`
			Price float64 `json:"price"`
		} `json:"variants"`
	}

	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "PROD001", response.Code)
	assert.Equal(t, 10.99, response.Price)
	assert.Len(t, response.Variants, 2)
	assert.Equal(t, 11.99, response.Variants[0].Price)
	assert.Equal(t, 10.99, response.Variants[1].Price)
}

func TestHandleGetProductDetails_NotFound(t *testing.T) {
	fakeRepo := &fakeProductsRepository{detailErr: gorm.ErrRecordNotFound}
	handler := NewCatalogHandler(fakeRepo, &fakeCategoriesRepository{})

	req := httptest.NewRequest(http.MethodGet, "/catalog/UNKNOWN", nil)
	recorder := httptest.NewRecorder()

	handler.HandleGetProductDetails(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestHandleGetCategories(t *testing.T) {
	fakeRepo := &fakeCategoriesRepository{
		categories: []models.Category{{Code: "clothing", Name: "Clothing"}},
	}
	handler := NewCatalogHandler(&fakeProductsRepository{}, fakeRepo)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	recorder := httptest.NewRecorder()

	handler.HandleCategories(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response struct {
		Categories []struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"categories"`
	}

	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response.Categories, 1)
	assert.Equal(t, "clothing", response.Categories[0].Code)
}

func TestHandleCreateCategory(t *testing.T) {
	fakeRepo := &fakeCategoriesRepository{}
	handler := NewCatalogHandler(&fakeProductsRepository{}, fakeRepo)

	body := strings.NewReader(`{"code":"new-category","name":"New Category"}`)
	req := httptest.NewRequest(http.MethodPost, "/categories", body)
	recorder := httptest.NewRecorder()

	handler.HandleCategories(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, 1, len(fakeRepo.categories))
	assert.Equal(t, "new-category", fakeRepo.categories[0].Code)
}

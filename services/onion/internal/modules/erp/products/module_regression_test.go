package products

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterAPIUsesPresenceAwareUpdateRoute(t *testing.T) {
	mux := http.NewServeMux()
	NewModule(context.Background(), nil).RegisterAPI(mux)

	updateRequest := httptest.NewRequest(http.MethodPut, "/erp/product/v1/products/1", nil)
	_, updatePattern := mux.Handler(updateRequest)
	if updatePattern != "PUT /erp/product/v1/products/{id}" {
		t.Fatalf("update route pattern = %q", updatePattern)
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/erp/product/v1/products", nil)
	_, listPattern := mux.Handler(listRequest)
	if listPattern != "/erp/product/v1/" {
		t.Fatalf("generated route pattern = %q, want /erp/product/v1/", listPattern)
	}
}

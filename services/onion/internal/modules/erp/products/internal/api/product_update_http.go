package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/platform/onekiterrors"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
)

// UpdateProductHTTP is the transport boundary for PUT product updates. The
// generated request type uses scalar optional fields, so the generated server
// cannot retain whether false, zero, or an empty image URL was sent. Decoding
// the raw object here preserves that presence before the domain merge.
func (a *ProductAPI) UpdateProductHTTP(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || productID <= 0 {
		writeProductUpdateHTTPError(w, onekiterrors.New(http.StatusBadRequest, "invalid product id"))
		return
	}

	body, present, err := decodeProductUpdateBody(w, r)
	if err != nil {
		writeProductUpdateHTTPError(w, onekiterrors.New(http.StatusBadRequest, "invalid request body"))
		return
	}
	if err := body.Validate(); err != nil {
		writeProductUpdateHTTPError(w, onekiterrors.New(http.StatusBadRequest, err.Error()))
		return
	}

	response, err := a.updateProduct(r.Context(), productID, productUpdateInput(body, present))
	if err != nil {
		writeProductUpdateHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func decodeProductUpdateBody(w http.ResponseWriter, r *http.Request) (*erpproductv1.UpdateProductBody, map[string]json.RawMessage, error) {
	if r.Body == nil {
		return nil, nil, errors.New("request body is required")
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}

	var body erpproductv1.UpdateProductBody
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, nil, err
	}
	var present map[string]json.RawMessage
	if err := json.Unmarshal(data, &present); err != nil {
		return nil, nil, err
	}
	if present == nil {
		present = make(map[string]json.RawMessage)
	}
	return &body, present, nil
}

func writeProductUpdateHTTPError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	var statusErr interface{ HTTPStatusCode() int }
	if errors.As(err, &statusErr) {
		candidate := statusErr.HTTPStatusCode()
		if candidate >= 100 && candidate <= 599 {
			status = candidate
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "internal server error"})
}

package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/suppliers/internal/application"
	"github.com/voxire/onion/internal/modules/suppliers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpsupplierv1 "github.com/voxire/voxie/gen/go/erp/supplier/v1"
)

type SupplierAPI struct {
	Service *application.Service
}

func NewSupplierAPI(service *application.Service) *SupplierAPI {
	return &SupplierAPI{Service: service}
}

func (a *SupplierAPI) ListSuppliers(ctx context.Context, req *erpsupplierv1.ListSuppliersRequest) (*erpsupplierv1.ListSuppliersResponse, error) {
	ownerID, err := supplierOwnerID(ctx, "erp:suppliers:view")
	if err != nil {
		return nil, err
	}
	page := req.GetPage()
	if page <= 0 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}

	suppliers, hasMore, listErr := a.Service.ListSuppliers(ctx, ownerID, int64(page), int64(pageSize))
	if listErr != nil {
		return nil, supplierStatusError(http.StatusInternalServerError, "Failed to list suppliers")
	}
	out := make([]*erpsupplierv1.Supplier, 0, len(suppliers))
	for _, s := range suppliers {
		out = append(out, newSupplierPayload(s))
	}
	return &erpsupplierv1.ListSuppliersResponse{
		Suppliers: out,
		HasMore:   hasMore,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

func (a *SupplierAPI) CreateSupplier(ctx context.Context, req *erpsupplierv1.CreateSupplierRequest) (*erpsupplierv1.CreateSupplierResponse, error) {
	ownerID, err := supplierOwnerID(ctx, "erp:suppliers:create")
	if err != nil {
		return nil, err
	}
	id, createErr := a.Service.CreateSupplier(ctx, ownerID, supplierInputFromCreate(req))
	if createErr != nil {
		return nil, supplierStatusError(http.StatusInternalServerError, "Failed to create supplier")
	}
	return &erpsupplierv1.CreateSupplierResponse{SupplierId: id, Message: "Supplier created successfully"}, nil
}

func (a *SupplierAPI) UpdateSupplier(ctx context.Context, req *erpsupplierv1.UpdateSupplierRequest) (*erpsupplierv1.UpdateSupplierResponse, error) {
	ownerID, err := supplierOwnerID(ctx, "erp:suppliers:edit")
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, supplierStatusError(http.StatusBadRequest, "Supplier body is required")
	}
	if updateErr := a.Service.UpdateSupplier(ctx, ownerID, req.GetId(), supplierInputFromUpdate(body)); updateErr != nil {
		return nil, supplierMutationError(updateErr)
	}
	return &erpsupplierv1.UpdateSupplierResponse{Message: "Supplier updated successfully"}, nil
}

func (a *SupplierAPI) DeleteSupplier(ctx context.Context, req *erpsupplierv1.DeleteSupplierRequest) (*erpsupplierv1.DeleteSupplierResponse, error) {
	ownerID, err := supplierOwnerID(ctx, "erp:suppliers:edit")
	if err != nil {
		return nil, err
	}
	if deleteErr := a.Service.DeleteSupplier(ctx, ownerID, req.GetId()); deleteErr != nil {
		return nil, supplierMutationError(deleteErr)
	}
	return &erpsupplierv1.DeleteSupplierResponse{Message: "Supplier deleted successfully"}, nil
}

func supplierOwnerID(ctx context.Context, permission string) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || !authPkg.HasAppPermission(authUser, authorization.AppERP, permission) {
		return 0, onekiterrors.PermissionDenied()
	}
	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return ownerID, nil
}

func supplierInputFromCreate(req *erpsupplierv1.CreateSupplierRequest) domain.SupplierInput {
	return domain.SupplierInput{
		Name:         strings.TrimSpace(req.GetName()),
		Email:        req.GetEmail(),
		Phone:        req.GetPhone(),
		Address:      req.GetAddress(),
		City:         req.GetCity(),
		Country:      req.GetCountry(),
		TaxID:        req.GetTaxId(),
		PaymentTerms: req.GetPaymentTerms(),
		Notes:        req.GetNotes(),
		IsActive:     req.GetIsActive(),
	}
}

func supplierInputFromUpdate(body *erpsupplierv1.UpdateSupplierBody) domain.SupplierInput {
	return domain.SupplierInput{
		Name:         strings.TrimSpace(body.GetName()),
		Email:        body.GetEmail(),
		Phone:        body.GetPhone(),
		Address:      body.GetAddress(),
		City:         body.GetCity(),
		Country:      body.GetCountry(),
		TaxID:        body.GetTaxId(),
		PaymentTerms: body.GetPaymentTerms(),
		Notes:        body.GetNotes(),
		IsActive:     body.GetIsActive(),
	}
}

func supplierStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}

func supplierMutationError(err error) error {
	if errors.Is(err, domain.ErrSupplierNotFound) {
		return supplierStatusError(http.StatusNotFound, "Supplier not found")
	}
	return supplierStatusError(http.StatusInternalServerError, "Failed to update supplier")
}

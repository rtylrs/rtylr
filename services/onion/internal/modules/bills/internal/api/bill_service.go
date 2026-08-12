package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/bills/internal/application"
	"github.com/voxire/onion/internal/modules/bills/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financebillv1 "github.com/voxire/voxie/gen/go/finance/bill/v1"
)

// listBillsPageSize bounds ListBills server-side. The voxie
// finance.bill.v1.ListBillsRequest contract has no pagination fields yet, so
// the endpoint returns at most one maximum-size page instead of every row.
const listBillsPageSize int32 = 500

type BillAPI struct {
	Service *application.Service
}

func NewBillAPI(service *application.Service) *BillAPI {
	return &BillAPI{Service: service}
}

func (a *BillAPI) ListBills(ctx context.Context, _ *financebillv1.ListBillsRequest) (*financebillv1.ListBillsResponse, error) {
	businessID, err := billBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	bills, listErr := a.Service.ListPaginated(ctx, businessID, listBillsPageSize, 0)
	if listErr != nil {
		return nil, billStatusError(http.StatusInternalServerError, "Failed to list bills")
	}
	out := make([]*financebillv1.Bill, 0, len(bills))
	for _, bill := range bills {
		out = append(out, newBillProto(bill))
	}
	return &financebillv1.ListBillsResponse{Bills: out}, nil
}

func (a *BillAPI) CreateBill(ctx context.Context, req *financebillv1.CreateBillRequest) (*financebillv1.CreateBillResponse, error) {
	businessID, err := billBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	input := domain.BillInput{
		SupplierID:  req.GetSupplierId(),
		Amount:      req.GetAmount(),
		Currency:    req.GetCurrency(),
		DueDate:     req.GetDueDate(),
		Status:      req.GetStatus(),
		Description: req.GetDescription(),
	}
	if validationErr := domain.ValidateBillInput(input); validationErr != nil {
		return nil, billStatusError(http.StatusBadRequest, validationErr.Error())
	}
	id, createErr := a.Service.Create(ctx, businessID, input)
	if createErr != nil {
		return nil, billStatusError(http.StatusInternalServerError, "Failed to create bill")
	}
	return &financebillv1.CreateBillResponse{Id: id, Message: "Bill created successfully"}, nil
}

func (a *BillAPI) UpdateBill(ctx context.Context, req *financebillv1.UpdateBillRequest) (*financebillv1.UpdateBillResponse, error) {
	businessID, err := billBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, billStatusError(http.StatusBadRequest, "Bill body is required")
	}
	input := domain.BillInput{
		SupplierID:  body.GetSupplierId(),
		Amount:      body.GetAmount(),
		Currency:    body.GetCurrency(),
		DueDate:     body.GetDueDate(),
		Status:      body.GetStatus(),
		Description: body.GetDescription(),
	}
	if validationErr := domain.ValidateBillInput(input); validationErr != nil {
		return nil, billStatusError(http.StatusBadRequest, validationErr.Error())
	}
	updateErr := a.Service.Update(ctx, businessID, req.GetId(), input)
	if updateErr != nil {
		if errors.Is(updateErr, domain.ErrBillNotFound) {
			return nil, billStatusError(http.StatusNotFound, "Bill not found")
		}
		logger.Ctx(ctx).Error("failed to update bill", logger.Any("bill_id", req.GetId()), logger.Error(updateErr))
		return nil, billStatusError(http.StatusInternalServerError, "Failed to update bill")
	}
	return &financebillv1.UpdateBillResponse{Message: "Bill updated successfully"}, nil
}

func (a *BillAPI) GetBill(ctx context.Context, req *financebillv1.GetBillRequest) (*financebillv1.GetBillResponse, error) {
	businessID, err := billBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	bill, getErr := a.Service.GetBill(ctx, businessID, req.GetId())
	if getErr != nil {
		if errors.Is(getErr, domain.ErrBillNotFound) {
			return nil, billStatusError(http.StatusNotFound, "Bill not found")
		}
		logger.Ctx(ctx).Error("failed to get bill", logger.Any("bill_id", req.GetId()), logger.Error(getErr))
		return nil, billStatusError(http.StatusInternalServerError, "Failed to get bill")
	}
	return &financebillv1.GetBillResponse{Bill: newBillProto(bill)}, nil
}

func (a *BillAPI) DeleteBill(ctx context.Context, req *financebillv1.DeleteBillRequest) (*financebillv1.DeleteBillResponse, error) {
	businessID, err := billBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	deleteErr := a.Service.DeleteBill(ctx, businessID, req.GetId())
	if deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrBillNotFound) {
			return nil, billStatusError(http.StatusNotFound, "Bill not found")
		}
		logger.Ctx(ctx).Error("failed to delete bill", logger.Any("bill_id", req.GetId()), logger.Error(deleteErr))
		return nil, billStatusError(http.StatusInternalServerError, "Failed to delete bill")
	}
	return &financebillv1.DeleteBillResponse{Message: "Bill deleted successfully"}, nil
}

func newBillProto(bill domain.Bill) *financebillv1.Bill {
	return &financebillv1.Bill{
		Id:          bill.ID,
		OwnerId:     bill.OwnerID,
		SupplierId:  bill.SupplierID,
		Amount:      bill.Amount,
		Currency:    bill.Currency,
		DueDate:     bill.DueDate.Format("2006-01-02"),
		Status:      bill.Status,
		Description: bill.Description,
		CreatedAt:   bill.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   bill.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func billBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return 0, billStatusError(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func billStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}

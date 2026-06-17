package handlers

import (
	"context"

	"github.com/voxire/onion/internal/modules/products/internal/application"
)

func productBoolDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func productFlagsForUpdate(
	ctx context.Context,
	service *application.Service,
	ownerID int64,
	productID int64,
	isActive *bool,
	trackInventory *bool,
) (bool, bool, error) {
	if isActive != nil && trackInventory != nil {
		return *isActive, *trackInventory, nil
	}

	current, err := service.GetProduct(ctx, ownerID, productID)
	if err != nil {
		return false, false, err
	}

	return productBoolDefault(isActive, current.IsActive),
		productBoolDefault(trackInventory, current.TrackInventory),
		nil
}

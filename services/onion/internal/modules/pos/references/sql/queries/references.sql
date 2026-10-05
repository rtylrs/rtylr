-- name: LocationExists :one
SELECT EXISTS(SELECT 1 FROM pos_locations WHERE business_id = ? AND id = ?) AS present;

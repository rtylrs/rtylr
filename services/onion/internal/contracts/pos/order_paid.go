package pos

const OrderPaidType = "pos_order_paid"

type PaidOrder struct {
	BranchID int64       `json:"branch_id"`
	Items    []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID int64   `json:"product_id"`
	Quantity  float64 `json:"quantity"`
}

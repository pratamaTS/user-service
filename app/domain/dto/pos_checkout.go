package dto

type POSCheckoutItem struct {
	ProductUUID string `json:"product_uuid"`
	Qty         int64  `json:"qty"`
}

type POSCheckoutRequest struct {
	BranchUUID string            `json:"branch_uuid"`
	Items      []POSCheckoutItem `json:"items"`

	Discount float64 `json:"discount"`
	Paid     float64 `json:"paid"`

	PaymentMethod string `json:"payment_method"` // CASH / TRANSFER / QRIS
	CreatedBy     string `json:"created_by"`
	Note          string `json:"note"`
}

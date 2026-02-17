package dao

type POSTransactionItem struct {
	ProductUUID string        `bson:"product_uuid" json:"product_uuid" validate:"required"`
	SKU         string        `bson:"sku" json:"sku"`
	Barcode     string        `bson:"barcode" json:"barcode"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	BaseUnit    string        `bson:"base_unit" json:"base_unit"`
	Units       []ProductUnit `bson:"units" json:"units"`
	Price       float64       `bson:"price" json:"price"`
	Qty         int64         `bson:"qty" json:"qty"`
	LineTotal   float64       `bson:"line_total" json:"line_total"`
}

type POSTransaction struct {
	BaseModel `bson:",inline"`

	BranchUUID string `bson:"branch_uuid" json:"branch_uuid"`

	// human readable receipt no (optional)
	ReceiptNo string `bson:"receipt_no" json:"receipt_no"`

	Items []POSTransactionItem `bson:"items" json:"items"`

	PaymentMethod string `bson:"payment_method" json:"payment_method"` // CASH / TRANSFER / QRIS

	SubTotal float64 `bson:"sub_total" json:"sub_total"`
	Discount float64 `bson:"discount" json:"discount"`
	Total    float64 `bson:"total" json:"total"`
	Paid     float64 `bson:"paid" json:"paid"`
	Change   float64 `bson:"change" json:"change"`

	Status      string `bson:"status" json:"status"` // PAID / VOID
	CreatedBy   string `bson:"created_by" json:"created_by"`
	VoidedBy    string `bson:"voided_by" json:"voided_by"`
	VoidedAt    int64  `bson:"voided_at" json:"voided_at"`
	VoidedAtStr string `bson:"voided_at_str" json:"voided_at_str"`
	Note        string `bson:"note" json:"note"`
}

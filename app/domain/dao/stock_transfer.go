package dao

type StockTransferStatus string

const (
	StockTransferInProgress StockTransferStatus = "IN_PROGRESS"
	StockTransferDone       StockTransferStatus = "DONE"
)

type StockTransferItem struct {
	ProductUUID string `bson:"product_uuid" json:"product_uuid"`
	SKU         string `bson:"sku,omitempty" json:"sku,omitempty"`
	Name        string `bson:"name,omitempty" json:"name,omitempty"`

	// input
	Unit string `bson:"unit" json:"unit"` // contoh: pcs/box
	Qty  int64  `bson:"qty" json:"qty"`   // qty sesuai unit

	// disimpan utk audit & proses stok
	ConversionToBase int64 `bson:"conversion_to_base" json:"conversion_to_base"` // contoh: box=12
	QtyBase          int64 `bson:"qty_base" json:"qty_base"`                     // qty * conversion
}

type StockTransfer struct {
	BaseModel      `bson:",inline"`
	FromBranchUUID string              `bson:"from_branch_uuid" json:"from_branch_uuid"`
	ToBranchUUID   string              `bson:"to_branch_uuid" json:"to_branch_uuid"`
	Notes          string              `bson:"notes,omitempty" json:"notes,omitempty"`
	Status         StockTransferStatus `bson:"status" json:"status"`

	Items []StockTransferItem `bson:"items" json:"items"`

	RequestedBy   string `bson:"requested_by" json:"requested_by"`
	ReceivedBy    string `bson:"received_by,omitempty" json:"received_by,omitempty"`
	ReceivedAt    int64  `bson:"received_at,omitempty" json:"received_at,omitempty"`
	ReceivedAtStr string `bson:"received_at_str,omitempty" json:"received_at_str,omitempty"`
}

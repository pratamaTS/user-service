package dao

import "go.mongodb.org/mongo-driver/bson"

type StockTransferStatus string

const (
	StockTransferPendingWarehouse StockTransferStatus = "PENDING_WAREHOUSE"
	StockTransferWaitingDriver    StockTransferStatus = "WAITING_DRIVER"
	StockTransferInProgress       StockTransferStatus = "IN_PROGRESS"
	StockTransferDone             StockTransferStatus = "DONE"
)

type StockTransferItem struct {
	ProductUUID string        `bson:"product_uuid" json:"product_uuid" validate:"required"`
	SKU         string        `bson:"sku" json:"sku"`
	Barcode     string        `bson:"barcode" json:"barcode"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	BaseUnit    string        `bson:"base_unit" json:"base_unit"`
	Units       []ProductUnit `bson:"units" json:"units"`
	Cost        float64       `bson:"cost" json:"cost"`
	Price       float64       `bson:"price" json:"price"`
	Qty         int64         `bson:"qty" json:"qty" validate:"required,min=1"`
	Image       string        `bson:"image" json:"image"`
}

type StockTransfer struct {
	BaseModel `bson:",inline"`

	FromBranchUUID string              `bson:"from_branch_uuid" json:"from_branch_uuid" validate:"required"`
	ToBranchUUID   string              `bson:"to_branch_uuid" json:"to_branch_uuid" validate:"required"`
	DriverUUID     string              `bson:"driver_uuid" json:"driver_uuid" validate:"required"`
	Status         StockTransferStatus `bson:"status" json:"status"`

	Items []StockTransferItem `bson:"items" json:"items" validate:"required,dive"`

	// audit
	RequestedBy   string `bson:"requested_by" json:"requested_by"`
	RequesterNote string `bson:"requester_note" json:"requester_note"`

	ApprovedBy   string `bson:"approved_by" json:"approved_by"`
	ApproverNote string `bson:"approver_note" json:"approver_note"`

	AcceptedBy   string `bson:"accepted_by" json:"accepted_by"`
	AccepterNote string `bson:"accepter_note" json:"accepter_note"`

	ReceivedBy   string `bson:"received_by" json:"received_by"`
	ReceiverNote string `bson:"receiver_note" json:"receiver_note"`

	ApprovedAt  int64  `bson:"approved_at" json:"approved_at"`
	AcceptedAt  int64  `bson:"accepted_at" json:"accepted_at"`
	ReceivedAt  int64  `bson:"received_at" json:"received_at"`
	ApprovedStr string `bson:"approved_at_str" json:"approved_at_str"`
	AcceptedStr string `bson:"accepted_at_str" json:"accepted_at_str"`
	ReceivedStr string `bson:"received_at_str" json:"received_at_str"`

	Driver bson.M `bson:"driver,omitempty" json:"driver,omitempty"`
}

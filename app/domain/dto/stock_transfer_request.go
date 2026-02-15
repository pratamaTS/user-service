package dto

type StockTransferCreateReq struct {
	FromBranchUUID string `json:"from_branch_uuid" validate:"required"`
	ToBranchUUID   string `json:"to_branch_uuid" validate:"required"`
	DriverUUID     string `json:"driver_uuid"`
	Notes          string `json:"notes"`
	RequestedBy    string `json:"requested_by"`
	Items          []struct {
		ProductUUID string `json:"product_uuid" validate:"required"`
		Qty         int64  `json:"qty" validate:"required,min=1"`
	} `json:"items" validate:"required,min=1,dive"`
}

type StockTransferWarehouseApproveReq struct {
	Notes string `json:"notes"`
}

type StockTransferDriverAcceptReq struct {
	Notes string `json:"notes"`
}

type StockTransferReceiveReq struct {
	Notes string `json:"notes"`
}

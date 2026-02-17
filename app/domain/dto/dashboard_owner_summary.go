package dto

type OwnerDashboardSummary struct {
	TotalProduct        int64   `json:"total_product"`
	TotalStockRequest   int64   `json:"total_stock_request"`
	StockRequestProcess int64   `json:"stock_request_process"`
	DriverAvailable     int64   `json:"driver_available"`
	JobWaitingAccept    int64   `json:"job_waiting_accept"`
	TransactionToday    int64   `json:"transaction_today"`
	RevenueMonth        float64 `json:"revenue_month"`
	LowStockSKU         int64   `json:"low_stock_sku"`
}

type OwnerDashboardRequest struct {
	LowStockThreshold int64 `json:"low_stock_threshold"`
}

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

// =========================
// KASIR
// =========================
type KasirDashboardSummary struct {
	TransactionToday      int64 `json:"transaction_today"`
	TransactionMonth      int64 `json:"transaction_month"`
	ProductInBranch       int64 `json:"product_in_branch"`
	StockRequestInProcess int64 `json:"stock_request_process"`
}

// =========================
// GUDANG
// =========================
type GudangDashboardSummary struct {
	TotalProduct           int64 `json:"total_product"`
	TotalStockRequestMonth int64 `json:"total_stock_request"`
	StockRequestInDriver   int64 `json:"stock_request_in_driver"`
	DriverAvailable        int64 `json:"driver_available"`
	ProductSentToBranch    int64 `json:"product_sent_to_branch"` // sum qty items terkirim
}

// =========================
// DRIVER
// =========================
type DriverDashboardSummary struct {
	JobToday   int64 `json:"job_today"`
	JobWaiting int64 `json:"job_waiting"`
}

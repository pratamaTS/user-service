package dto

type POSHistoryItem struct {
	UUID          string `json:"uuid"`
	BranchUUID    string `json:"branch_uuid"`
	ReceiptNo     string `json:"receipt_no"`
	Status        string `json:"status"` // PAID / VOID
	PaymentMethod string `json:"payment_method"`

	SubTotal float64 `json:"sub_total"`
	Discount float64 `json:"discount"`
	Total    float64 `json:"total"`
	Paid     float64 `json:"paid"`
	Change   float64 `json:"change"`

	CreatedBy string `json:"created_by"`
	VoidedBy  string `json:"voided_by"`
	Note      string `json:"note"`

	// ✅ formatted time (dd-MM-YYYY hh:mm:ss)
	TrxAt  string `json:"trx_at"`
	VoidAt string `json:"void_at"`
}

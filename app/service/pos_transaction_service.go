package service

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
	"harjonan.id/user-service/app/repository"
)

type POSTransactionService interface {
	Checkout(ctx *gin.Context)
	Detail(ctx *gin.Context)
	List(ctx *gin.Context)
	Void(ctx *gin.Context)
	ScanByBarcode(ctx *gin.Context)
}

type POSTransactionServiceImpl struct {
	trxRepo  repository.POSTransactionRepository
	prodRepo repository.ProductRepository
}

func NewPOSTransactionService(trxRepo repository.POSTransactionRepository, prodRepo repository.ProductRepository) *POSTransactionServiceImpl {
	return &POSTransactionServiceImpl{trxRepo: trxRepo, prodRepo: prodRepo}
}

// body: { "branch_uuid": "...", "barcode": "..." }
func (s *POSTransactionServiceImpl) ScanByBarcode(ctx *gin.Context) {
	var req struct {
		BranchUUID string `json:"branch_uuid"`
		Barcode    string `json:"barcode"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	req.BranchUUID = strings.TrimSpace(req.BranchUUID)
	req.Barcode = strings.TrimSpace(req.Barcode)
	if req.BranchUUID == "" || req.Barcode == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("branch_uuid & barcode required"))
		return
	}

	fr := dto.FilterRequest{
		Search: "",
		SortBy: map[string]any{"created_at": "desc"},
		FilterBy: map[string]any{
			"branch_uuid": req.BranchUUID,
			"barcode":     req.Barcode,
			"is_active":   true,
		},
	}
	fr.Pagination.Page = 1
	fr.Pagination.PageSize = 1

	list, err := s.prodRepo.ListProduct(&fr)
	if err != nil || len(list) == 0 {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, errors.New("barcode not found"))
		return
	}
	helpers.JsonOK(ctx, "success", list[0])
}

func (s *POSTransactionServiceImpl) Checkout(ctx *gin.Context) {
	var req dto.POSCheckoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}

	req.BranchUUID = strings.TrimSpace(req.BranchUUID)
	req.PaymentMethod = strings.TrimSpace(strings.ToUpper(req.PaymentMethod))

	if req.BranchUUID == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("branch_uuid required"))
		return
	}
	if len(req.Items) == 0 {
		helpers.JsonErr[any](ctx, "missing items", http.StatusBadRequest, errors.New("items required"))
		return
	}

	// payment method (category only)
	if req.PaymentMethod == "" {
		req.PaymentMethod = "CASH" // default
	}
	switch req.PaymentMethod {
	case "CASH", "TRANSFER", "QRIS":
	default:
		helpers.JsonErr[any](ctx, "invalid payment_method", http.StatusBadRequest, errors.New("payment_method must be CASH/TRANSFER/QRIS"))
		return
	}

	if req.Discount < 0 {
		req.Discount = 0
	}
	if req.Paid < 0 {
		req.Paid = 0
	}

	// build items from DB (server-side pricing + validate stock)
	items := make([]dao.POSTransactionItem, 0, len(req.Items))
	var subTotal float64

	type decPlan struct {
		ProductUUID string
		Qty         int64
	}
	var plans []decPlan

	for _, it := range req.Items {
		if strings.TrimSpace(it.ProductUUID) == "" || it.Qty <= 0 {
			helpers.JsonErr[any](ctx, "invalid item", http.StatusBadRequest, errors.New("product_uuid & qty required"))
			return
		}

		p, err := s.prodRepo.DetailProduct(it.ProductUUID)
		if err != nil {
			helpers.JsonErr[any](ctx, "product not found", http.StatusNotFound, err)
			return
		}
		if p.BranchUUID != req.BranchUUID {
			helpers.JsonErr[any](ctx, "branch mismatch", http.StatusBadRequest, errors.New("product branch_uuid mismatch"))
			return
		}
		if !p.IsActive {
			helpers.JsonErr[any](ctx, "inactive product", http.StatusBadRequest, errors.New("product inactive"))
			return
		}
		if p.Stock < it.Qty {
			helpers.JsonErr[any](ctx, "stock not enough", http.StatusBadRequest, errors.New(p.Name+" stock not enough"))
			return
		}

		line := float64(it.Qty) * p.Price
		subTotal += line

		items = append(items, dao.POSTransactionItem{
			ProductUUID: p.UUID,
			SKU:         p.SKU,
			Barcode:     p.Barcode,
			Name:        p.Name,
			Description: p.Description,
			BaseUnit:    p.BaseUnit,
			Units:       p.Units,
			Price:       p.Price,
			Qty:         it.Qty,
			LineTotal:   line,
		})
		plans = append(plans, decPlan{ProductUUID: p.UUID, Qty: it.Qty})
	}

	total := subTotal - req.Discount
	if total < 0 {
		total = 0
	}
	if req.Paid < total {
		helpers.JsonErr[any](ctx, "payment not enough", http.StatusBadRequest, errors.New("paid < total"))
		return
	}
	change := req.Paid - total

	// 1) decrement stock per item (atomic per document)
	decOK := make([]decPlan, 0, len(plans))
	for _, pl := range plans {
		err := s.prodRepo.DecreaseStockIfEnough(pl.ProductUUID, pl.Qty)
		if err != nil {
			for i := len(decOK) - 1; i >= 0; i-- {
				_ = s.prodRepo.IncreaseStock(decOK[i].ProductUUID, decOK[i].Qty)
			}
			helpers.JsonErr[any](ctx, "failed to update stock", http.StatusInternalServerError, err)
			return
		}
		decOK = append(decOK, pl)
	}

	// 2) insert transaction
	now := time.Now()
	ms := now.UnixNano() / 1e6

	trx := dao.POSTransaction{
		BranchUUID:    req.BranchUUID,
		ReceiptNo:     "TRX-" + strconv.FormatInt(ms, 10),
		PaymentMethod: req.PaymentMethod,
		Items:         items,
		SubTotal:      subTotal,
		Discount:      req.Discount,
		Total:         total,
		Paid:          req.Paid,
		Change:        change,
		Status:        "PAID",
		CreatedBy:     req.CreatedBy,
		Note:          req.Note,
	}

	out, err := s.trxRepo.Insert(&trx)
	if err != nil {
		// rollback stock kalau insert gagal
		for i := len(decOK) - 1; i >= 0; i-- {
			_ = s.prodRepo.IncreaseStock(decOK[i].ProductUUID, decOK[i].Qty)
		}
		helpers.JsonErr[any](ctx, "failed to checkout", http.StatusInternalServerError, err)
		return
	}

	helpers.JsonOK(ctx, "success", out)
}

func (s *POSTransactionServiceImpl) Detail(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if strings.TrimSpace(uuid) == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	data, err := s.trxRepo.Detail(uuid)
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *POSTransactionServiceImpl) List(ctx *gin.Context) {
	var req dto.FilterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	data, err := s.trxRepo.List(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list transaction", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

// endpoint: POST /pos/transactions/:uuid/void  body: { "voided_by":"...", "note":"..." }
func (s *POSTransactionServiceImpl) Void(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if strings.TrimSpace(uuid) == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}

	var req struct {
		VoidedBy string `json:"voided_by"`
		Note     string `json:"note"`
	}
	_ = ctx.ShouldBindJSON(&req)

	trx, err := s.trxRepo.Detail(uuid)
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}
	if trx.Status == "VOID" {
		helpers.JsonOK(ctx, "success", trx)
		return
	}

	// return stock
	for _, it := range trx.Items {
		if strings.TrimSpace(it.ProductUUID) == "" || it.Qty <= 0 {
			continue
		}
		err = s.prodRepo.IncreaseStock(it.ProductUUID, it.Qty)
		if err != nil {
			helpers.JsonErr[any](ctx, "failed to return stock", http.StatusInternalServerError, err)
			return
		}
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	out, err := s.trxRepo.UpdateStatus(uuid, "VOID", bson.M{
		"voided_by":     req.VoidedBy,
		"voided_at":     now.Unix(),
		"voided_at_str": nowStr,
		"note":          req.Note,
	})
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to void", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", out)
}

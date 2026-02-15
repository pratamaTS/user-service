package service

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
	"harjonan.id/user-service/app/repository"
)

type ProductService interface {
	Upsert(ctx *gin.Context)
	Detail(ctx *gin.Context)
	List(ctx *gin.Context)
	Delete(ctx *gin.Context)
	BulkUpload(ctx *gin.Context)
}

type ProductServiceImpl struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductServiceImpl {
	return &ProductServiceImpl{repo: repo}
}

func (s *ProductServiceImpl) Upsert(ctx *gin.Context) {
	var req dao.Product
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	if req.BranchUUID == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("branch_uuid required"))
		return
	}
	if req.Name == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("name required"))
		return
	}
	if req.BaseUnit == "" {
		helpers.JsonErr[any](ctx, "missing identifier", http.StatusBadRequest, errors.New("base_unit required"))
		return
	}

	res, err := s.repo.SaveProduct(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to save product", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", res)
}

func (s *ProductServiceImpl) Detail(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	data, err := s.repo.DetailProduct(uuid)
	if err != nil {
		helpers.JsonErr[any](ctx, "not found", http.StatusNotFound, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *ProductServiceImpl) List(ctx *gin.Context) {
	var req dto.APIRequest[dto.FilterRequest]
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JsonErr[any](ctx, "invalid request", http.StatusBadRequest, err)
		return
	}
	data, err := s.repo.ListProduct(&req)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to list product", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK(ctx, "success", data)
}

func (s *ProductServiceImpl) Delete(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		helpers.JsonErr[any](ctx, "missing uuid", http.StatusBadRequest, errors.New("uuid required"))
		return
	}
	if err := s.repo.DeleteProduct(uuid); err != nil {
		helpers.JsonErr[any](ctx, "failed to delete product", http.StatusInternalServerError, err)
		return
	}
	helpers.JsonOK[struct{}](ctx, "success", struct{}{})
}

// ---------- BULK UPLOAD ----------
//
// Template kolom (CSV/XLSX) (header WAJIB match salah satu):
// branch_uuid, sku, barcode, name, description, base_unit, units, cost, price, is_active
//
// Format units:
// pcs:1|box:12|karton:240
// (artinya: base_unit misal pcs, box=12 pcs, karton=240 pcs)
//
// is_active: true/false/1/0/yes/no
func (s *ProductServiceImpl) BulkUpload(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		helpers.JsonErr[any](ctx, "missing file", http.StatusBadRequest, err)
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	f, err := file.Open()
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to open file", http.StatusBadRequest, err)
		return
	}
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		helpers.JsonErr[any](ctx, "failed to read file", http.StatusBadRequest, err)
		return
	}

	var products []dao.Product
	switch ext {
	case ".csv":
		products, err = parseProductsCSV(bytes.NewReader(raw))
	case ".xlsx":
		products, err = parseProductsXLSX(bytes.NewReader(raw))
	default:
		helpers.JsonErr[any](ctx, "unsupported file", http.StatusBadRequest, errors.New("only .csv or .xlsx supported"))
		return
	}

	if err != nil {
		helpers.JsonErr[any](ctx, "invalid file content", http.StatusBadRequest, err)
		return
	}
	if len(products) == 0 {
		helpers.JsonErr[any](ctx, "empty data", http.StatusBadRequest, errors.New("no rows found"))
		return
	}

	ok, fail, errs := s.repo.BulkUpsertProducts(products)

	resp := gin.H{
		"success_count": ok,
		"fail_count":    fail,
		"errors": func() []string {
			out := make([]string, 0, len(errs))
			for _, e := range errs {
				out = append(out, e.Error())
			}
			return out
		}(),
	}
	helpers.JsonOK(ctx, "success", resp)
}

// ---------- helpers parse ----------

func parseProductsCSV(r io.Reader) ([]dao.Product, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true

	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, nil
	}

	header := normalizeHeader(rows[0])
	idx := indexMap(header)

	var out []dao.Product
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if isRowEmpty(row) {
			continue
		}
		p, err := mapRowToProduct(func(col string) string {
			j, ok := idx[col]
			if !ok || j >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[j])
		})
		if err != nil {
			return nil, errors.New("row " + strconv.Itoa(i+1) + ": " + err.Error())
		}
		out = append(out, p)
	}
	return out, nil
}

func parseProductsXLSX(r io.Reader) ([]dao.Product, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("no sheet found")
	}
	sheet := sheets[0]

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, nil
	}

	header := normalizeHeader(rows[0])
	idx := indexMap(header)

	var out []dao.Product
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if isRowEmpty(row) {
			continue
		}
		p, err := mapRowToProduct(func(col string) string {
			j, ok := idx[col]
			if !ok || j >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[j])
		})
		if err != nil {
			return nil, errors.New("row " + strconv.Itoa(i+1) + ": " + err.Error())
		}
		out = append(out, p)
	}
	return out, nil
}

func mapRowToProduct(get func(col string) string) (dao.Product, error) {
	p := dao.Product{}
	p.BranchUUID = get("branch_uuid")
	p.SKU = get("sku")
	p.Barcode = get("barcode")
	p.Name = get("name")
	p.Description = get("description")
	p.BaseUnit = get("base_unit")
	p.Units = parseUnits(get("units"))
	p.Cost = parseFloat(get("cost"))
	p.Price = parseFloat(get("price"))
	p.IsActive = parseBool(get("is_active"), true)

	// validation minimal
	if p.BranchUUID == "" {
		return dao.Product{}, errors.New("branch_uuid required")
	}
	if p.Name == "" {
		return dao.Product{}, errors.New("name required")
	}
	if p.BaseUnit == "" {
		return dao.Product{}, errors.New("base_unit required")
	}

	// default: selalu pastikan base unit juga ada di Units minimal conversion 1
	p.Units = ensureBaseUnit(p.BaseUnit, p.Units)

	return p, nil
}

func normalizeHeader(h []string) []string {
	out := make([]string, 0, len(h))
	for _, v := range h {
		v = strings.TrimSpace(strings.ToLower(v))
		v = strings.ReplaceAll(v, " ", "_")
		out = append(out, v)
	}
	return out
}

func indexMap(header []string) map[string]int {
	m := map[string]int{}
	for i, col := range header {
		if col == "" {
			continue
		}
		m[col] = i
	}
	return m
}

func isRowEmpty(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func parseUnits(s string) []dao.ProductUnit {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	// format: pcs:1|box:12|karton:240
	parts := strings.Split(s, "|")
	var out []dao.ProductUnit
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.Split(part, ":")
		name := strings.TrimSpace(kv[0])
		if name == "" {
			continue
		}
		conv := 1.0
		if len(kv) >= 2 {
			conv = parseFloat(strings.TrimSpace(kv[1]))
			if conv <= 0 {
				conv = 1
			}
		}
		out = append(out, dao.ProductUnit{Name: name, ConversionToBase: conv})
	}
	return out
}

func ensureBaseUnit(base string, units []dao.ProductUnit) []dao.ProductUnit {
	base = strings.TrimSpace(base)
	if base == "" {
		return units
	}
	for _, u := range units {
		if strings.EqualFold(u.Name, base) {
			// pastikan conversion 1
			if u.ConversionToBase <= 0 {
				u.ConversionToBase = 1
			}
			return units
		}
	}
	return append([]dao.ProductUnit{{Name: base, ConversionToBase: 1}}, units...)
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// amanin format "200,000" -> "200000"
	s = strings.ReplaceAll(s, ",", "")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func parseBool(s string, def bool) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return def
	}
	switch s {
	case "1", "true", "yes", "y", "aktif", "active":
		return true
	case "0", "false", "no", "n", "nonaktif", "inactive":
		return false
	default:
		return def
	}
}

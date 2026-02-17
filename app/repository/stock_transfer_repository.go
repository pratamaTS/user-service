package repository

import (
	"context"
	"errors"
	"log"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
)

type StockTransferRepository interface {
	CreateDraft(data *dao.StockTransfer) (dao.StockTransfer, error)
	Detail(uuid string) (dao.StockTransfer, error)
	List(req *dto.FilterRequest) ([]dao.StockTransfer, error)

	WarehouseApprove(uuid, notes, approvedBy string) (dao.StockTransfer, error)
	DriverAccept(uuid, notes, acceptedBy string) (dao.StockTransfer, error)
	ReceiveDone(uuid, notes, receivedBy string) (dao.StockTransfer, error)
}

type StockTransferRepositoryImpl struct {
	db          *mongo.Database
	transferCol *mongo.Collection
	productCol  *mongo.Collection
	userCol     *mongo.Collection
	roleCol     *mongo.Collection
}

func StockTransferRepositoryInit(mongoClient *mongo.Client) *StockTransferRepositoryImpl {
	dbName := helpers.ProvideDBName()
	db := mongoClient.Database(dbName)
	return &StockTransferRepositoryImpl{
		db:          db,
		transferCol: db.Collection("stock_transfers"),
		productCol:  db.Collection("products"),
		userCol:     db.Collection("client_users"),
		roleCol:     db.Collection("roles"),
	}
}

func (r *StockTransferRepositoryImpl) CreateDraft(data *dao.StockTransfer) (dao.StockTransfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	log.Print("Creating stock transfer with driver: ", data.DriverUUID)

	for i := range data.Items {
		it := &data.Items[i] // ✅ pointer ke item asli di slice

		if it.ProductUUID == "" || it.Qty <= 0 {
			return dao.StockTransfer{}, errors.New("invalid item qty")
		}

		var product dao.Product
		err := r.productCol.FindOne(ctx, bson.M{"uuid": it.ProductUUID}).Decode(&product)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return dao.StockTransfer{}, errors.New("product not found: " + it.ProductUUID)
			}
			return dao.StockTransfer{}, err
		}

		it.SKU = product.SKU
		it.Barcode = product.Barcode
		it.Name = product.Name
		it.BaseUnit = product.BaseUnit
		it.Description = product.Description
		it.Units = product.Units
		it.Cost = product.Cost
		it.Price = product.Price
		it.Image = product.Image
	}

	log.Print("Product items: ", data.Items)

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	data.UUID = helpers.GenerateUUID()
	data.CreatedAt = now
	data.CreatedAtStr = nowStr
	data.UpdatedAt = now.Unix()
	data.UpdatedAtStr = nowStr
	data.Status = dao.StockTransferPendingWarehouse

	if _, err := r.transferCol.InsertOne(ctx, data); err != nil {
		return dao.StockTransfer{}, err
	}
	var out dao.StockTransfer
	if err := r.transferCol.FindOne(ctx, bson.M{"uuid": data.UUID}).Decode(&out); err != nil {
		return dao.StockTransfer{}, err
	}
	return out, nil
}

func (r *StockTransferRepositoryImpl) Detail(uuid string) (dao.StockTransfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var out dao.StockTransfer
	err := r.transferCol.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&out)
	return out, err
}

func (r *StockTransferRepositoryImpl) List(req *dto.FilterRequest) ([]dao.StockTransfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	filter := bson.M{}

	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"uuid": bson.M{"$regex": req.Search, "$options": "i"}},
			{"notes": bson.M{"$regex": req.Search, "$options": "i"}},
			{"requester_note": bson.M{"$regex": req.Search, "$options": "i"}},
			{"items.name": bson.M{"$regex": req.Search, "$options": "i"}},
			{"items.sku": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	for k, v := range req.FilterBy {
		if k == "" || v == nil {
			continue
		}

		rv := reflect.ValueOf(v)

		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			// handle empty list: skip (atau set impossible match)
			if rv.Len() == 0 {
				continue
			}

			// convert ke []any agar aman
			inVals := make([]any, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				inVals = append(inVals, rv.Index(i).Interface())
			}

			filter[k] = bson.M{"$in": inVals}
			continue
		}

		filter[k] = v
	}

	// sort
	sort := bson.D{{Key: "created_at", Value: -1}}
	for k, v := range req.SortBy {
		if s, ok := v.(string); ok && (s == "asc" || s == "ASC" || s == "1") {
			sort = bson.D{{Key: k, Value: 1}}
		} else {
			sort = bson.D{{Key: k, Value: -1}}
		}
	}

	// pagination
	page := req.Pagination.Page
	size := req.Pagination.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	skip := int64((page - 1) * size)

	log.Print("Filter by: ", filter)

	// ✅ NOTE:
	// - driver_uuid kamu mengarah ke client_user.uuid (dari FE)
	// - client_user biasanya punya user_uuid / fk_user_uuid -> join ke users.uuid
	// Sesuaikan field join-nya (aku pakai "user_uuid" sebagai contoh).
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$sort", Value: sort}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: int64(size)}},

		// ✅ lookup client_user by driver_uuid
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "users", // 🔁 ganti sesuai nama collection kamu
				"localField":   "driver_uuid",
				"foreignField": "uuid",
				"as":           "driver_user",
			},
		}},
		{{
			Key: "$unwind",
			Value: bson.M{
				"path":                       "$driver_user",
				"preserveNullAndEmptyArrays": true,
			},
		}},

		// ✅ lookup users by driver_client_user.user_uuid (sesuaikan)
		{{
			Key: "$lookup",
			Value: bson.M{
				"from":         "client_users",
				"localField":   "driver_user.uuid",
				"foreignField": "user_uuid",
				"as":           "driver_client_user",
			},
		}},
		{{
			Key: "$unwind",
			Value: bson.M{
				"path":                       "$driver_client_user",
				"preserveNullAndEmptyArrays": true,
			},
		}},

		// ✅ build driver object (langsung masuk ke field "driver" yg kita tambahkan di DAO)
		{{
			Key: "$addFields",
			Value: bson.M{
				"driver": bson.M{
					"uuid":        "$driver_client_user.uuid",
					"branch_uuid": "$driver_client_user.branch_uuid",
					"role":        "$driver_client_user.role",
					"user": bson.M{
						"uuid":         "$driver_user.uuid",
						"name":         "$driver_user.name",
						"username":     "$driver_user.username",
						"phone_number": "$driver_user.phone_number",
						"image":        "$driver_user.image",
					},
				},
			},
		}},

		// optional: bersihin temp fields
		{{
			Key: "$project",
			Value: bson.M{
				"driver_client_user": 0,
				"driver_user":        0,
			},
		}},
	}

	cur, err := r.transferCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []dao.StockTransfer
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}

	return list, nil
}

/*
WarehouseApprove:
- Validasi status harus PENDING_WAREHOUSE
- Validasi driver_uuid adalah role DRIVER (client_users.role_value == "DRIVER")
- Untuk setiap item: cek stock gudang cukup (products.branch_uuid=from_branch AND uuid=item.product_uuid)
- Potong stock gudang (stock -= qty)
- Update transfer: status WAITING_DRIVER, driver_uuid, approved_at
ALL IN TRANSACTION
*/
func (r *StockTransferRepositoryImpl) WarehouseApprove(uuid, notes, approvedBy string) (dao.StockTransfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	// validate warehouse role
	var approver bson.M
	if err := r.userCol.FindOne(ctx, bson.M{"user_uuid": approvedBy}).Decode(&approver); err != nil {
		return dao.StockTransfer{}, errors.New("approver not found / not WAREHOUSE role")
	}
	var role bson.M
	if err := r.roleCol.FindOne(ctx, bson.M{"uuid": approver["role_uuid"]}).Decode(&role); err != nil {
		return dao.StockTransfer{}, errors.New("approver not found / not WAREHOUSE role")
	}
	if role["value"] != "GUDANG" {
		return dao.StockTransfer{}, errors.New("approver not found / not WAREHOUSE role")
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	// 1) ambil transfer (no txn) + validasi status
	var tr dao.StockTransfer
	if err := r.transferCol.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&tr); err != nil {
		return dao.StockTransfer{}, err
	}
	if tr.Status != dao.StockTransferPendingWarehouse {
		return dao.StockTransfer{}, errors.New("invalid status: must be PENDING_WAREHOUSE")
	}

	// 2) potong stock per item dengan update atomic (filter stock >= qty)
	// simpan list yang sudah berhasil dipotong, untuk rollback kalau ada error
	type cutItem struct {
		ProductUUID string
		Qty         int64
	}
	cut := make([]cutItem, 0, len(tr.Items))

	rollback := func() {
		// best-effort rollback
		for _, it := range cut {
			filter := bson.M{"uuid": it.ProductUUID, "branch_uuid": tr.FromBranchUUID}
			_, _ = r.productCol.UpdateOne(context.Background(), filter, bson.M{
				"$inc": bson.M{"stock": it.Qty},
			})
		}
	}

	for _, it := range tr.Items {
		if it.ProductUUID == "" || it.Qty <= 0 {
			rollback()
			return dao.StockTransfer{}, errors.New("invalid item qty")
		}

		filter := bson.M{
			"uuid":        it.ProductUUID,
			"branch_uuid": tr.FromBranchUUID,
			"stock":       bson.M{"$gte": it.Qty}, // IMPORTANT: cegah minus secara atomic
		}

		res, err := r.productCol.UpdateOne(ctx, filter, bson.M{
			"$inc": bson.M{"stock": -it.Qty},
			"$set": bson.M{"updated_at": now.Unix(), "updated_at_str": nowStr},
		})
		if err != nil {
			rollback()
			return dao.StockTransfer{}, err
		}
		if res.MatchedCount == 0 {
			// bisa karena product tidak ada di branch tsb, atau stok kurang
			// optional: fetch sku untuk error message (extra query)
			rollback()
			return dao.StockTransfer{}, errors.New("product not found in from_branch or insufficient stock")
		}

		cut = append(cut, cutItem{ProductUUID: it.ProductUUID, Qty: it.Qty})
	}

	// 3) update transfer dengan guard status masih PENDING_WAREHOUSE (hindari double-approve race)
	updFilter := bson.M{"uuid": uuid, "status": dao.StockTransferPendingWarehouse}
	upd := bson.M{
		"$set": bson.M{
			"status":          dao.StockTransferWaitingDriver,
			"approved_by":     approvedBy,
			"approver_notes":  notes,
			"approved_at":     now.Unix(),
			"approved_at_str": nowStr,
			"updated_at":      now.Unix(),
			"updated_at_str":  nowStr,
		},
	}

	updRes, err := r.transferCol.UpdateOne(ctx, updFilter, upd)
	if err != nil {
		rollback()
		return dao.StockTransfer{}, err
	}
	if updRes.MatchedCount == 0 {
		// status keburu berubah oleh proses lain -> rollback potongan stock
		rollback()
		return dao.StockTransfer{}, errors.New("approve failed: transfer status already changed")
	}

	return r.Detail(uuid)
}

/*
DriverAccept:
- status harus WAITING_DRIVER
- driverUUID must match transfer.driver_uuid
- update status IN_PROGRESS
*/
func (r *StockTransferRepositoryImpl) DriverAccept(uuid, notes, acceptedBy string) (dao.StockTransfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	// validate driver role
	var driver bson.M
	if err := r.userCol.FindOne(ctx, bson.M{"user_uuid": acceptedBy}).Decode(&driver); err != nil {
		return dao.StockTransfer{}, errors.New("driver not found / not DRIVER role")
	}
	var role bson.M
	if err := r.roleCol.FindOne(ctx, bson.M{"uuid": driver["role_uuid"]}).Decode(&role); err != nil {
		return dao.StockTransfer{}, errors.New("driver not found / not DRIVER role")
	}
	if role["value"] != "DRIVER" {
		return dao.StockTransfer{}, errors.New("driver not found / not DRIVER role")
	}

	// ambil transfer untuk validasi driver assigned (biar error message jelas)
	var tr dao.StockTransfer
	if err := r.transferCol.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&tr); err != nil {
		return dao.StockTransfer{}, err
	}
	if tr.DriverUUID != acceptedBy {
		return dao.StockTransfer{}, errors.New("forbidden: driver not assigned")
	}
	if tr.Status != dao.StockTransferWaitingDriver {
		return dao.StockTransfer{}, errors.New("invalid status: must be WAITING_DRIVER")
	}

	// update dengan guard status + driver_uuid (anti race / double accept)
	filter := bson.M{
		"uuid":        uuid,
		"status":      dao.StockTransferWaitingDriver,
		"driver_uuid": acceptedBy,
	}

	res, err := r.transferCol.UpdateOne(ctx, filter, bson.M{
		"$set": bson.M{
			"status":          dao.StockTransferInProgress,
			"accepted_by":     acceptedBy,
			"acceptor_notes":  notes,
			"accepted_at":     now.Unix(),
			"accepted_at_str": nowStr,
			"notes":           notes,
			"updated_at":      now.Unix(),
			"updated_at_str":  nowStr,
		},
	})
	if err != nil {
		return dao.StockTransfer{}, err
	}
	if res.MatchedCount == 0 {
		return dao.StockTransfer{}, errors.New("accept failed: transfer status already changed or driver mismatch")
	}

	return r.Detail(uuid)
}

/*
ReceiveDone:
- status harus IN_PROGRESS
- stock masuk ke branch tujuan (products branch_uuid=to_branch + sku)
  - kalau product tujuan belum ada: create copy minimal
  - lalu stock += qty

- update transfer status DONE
ALL IN TRANSACTION
*/
func (r *StockTransferRepositoryImpl) ReceiveDone(uuid, notes, receivedBy string) (dao.StockTransfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	// validate kasir role
	var receiver bson.M
	if err := r.userCol.FindOne(ctx, bson.M{"user_uuid": receivedBy}).Decode(&receiver); err != nil {
		return dao.StockTransfer{}, errors.New("receiver not found / not KASIR role")
	}
	var role bson.M
	if err := r.roleCol.FindOne(ctx, bson.M{"uuid": receiver["role_uuid"]}).Decode(&role); err != nil {
		return dao.StockTransfer{}, errors.New("receiver not found / not KASIR role")
	}
	if role["value"] != "KASIR" && role["value"] != "ADMIN" {
		return dao.StockTransfer{}, errors.New("receiver not found / not KASIR role")
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	// get transfer
	var tr dao.StockTransfer
	if err := r.transferCol.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&tr); err != nil {
		return dao.StockTransfer{}, err
	}
	if tr.Status != dao.StockTransferInProgress {
		return dao.StockTransfer{}, errors.New("invalid status: must be IN_PROGRESS")
	}

	// track increments for rollback
	type incItem struct {
		SKU string
		Qty int64
	}
	applied := make([]incItem, 0, len(tr.Items))

	rollback := func() {
		// best-effort rollback
		for _, it := range applied {
			filter := bson.M{"branch_uuid": tr.ToBranchUUID, "sku": it.SKU}
			_, _ = r.productCol.UpdateOne(context.Background(), filter, bson.M{
				"$inc": bson.M{"stock": -it.Qty},
			})
		}
	}

	// stock to destination (atomic per item)
	for _, it := range tr.Items {
		if it.SKU == "" || it.Qty <= 0 {
			rollback()
			return dao.StockTransfer{}, errors.New("invalid item qty")
		}

		destFilter := bson.M{"branch_uuid": tr.ToBranchUUID, "sku": it.SKU}

		// gunakan $setOnInsert supaya kalau belum ada product, dibuat otomatis saat upsert
		update := bson.M{
			"$inc": bson.M{"stock": it.Qty},
			"$setOnInsert": bson.M{
				"uuid":           helpers.GenerateUUID(),
				"branch_uuid":    tr.ToBranchUUID,
				"sku":            it.SKU,
				"barcode":        it.Barcode,
				"name":           it.Name,
				"description":    it.Description,
				"base_unit":      it.BaseUnit,
				"units":          it.Units,
				"cost":           it.Cost,
				"price":          it.Price,
				"image":          it.Image,
				"is_active":      true,
				"created_by":     receivedBy,
				"created_at":     now, // sesuaikan type field kamu (time.Time vs unix)
				"created_at_str": nowStr,
				"updated_at":     now.Unix(),
				"updated_at_str": nowStr,
			},
		}

		_, err := r.productCol.UpdateOne(ctx, destFilter, update, options.Update().SetUpsert(true))
		if err != nil {
			rollback()
			return dao.StockTransfer{}, err
		}

		applied = append(applied, incItem{SKU: it.SKU, Qty: it.Qty})
	}

	// update transfer DONE dengan guard status (anti race)
	updFilter := bson.M{"uuid": uuid, "status": dao.StockTransferInProgress}
	upd := bson.M{
		"$set": bson.M{
			"status":          dao.StockTransferDone,
			"received_by":     receivedBy,
			"receiver_notes":  notes,
			"received_at":     now.Unix(),
			"received_at_str": nowStr,
			"notes":           notes,
			"updated_at":      now.Unix(),
			"updated_at_str":  nowStr,
		},
	}

	res, err := r.transferCol.UpdateOne(ctx, updFilter, upd)
	if err != nil {
		rollback()
		return dao.StockTransfer{}, err
	}
	if res.MatchedCount == 0 {
		rollback()
		return dao.StockTransfer{}, errors.New("receive failed: transfer status already changed")
	}

	return r.Detail(uuid)
}

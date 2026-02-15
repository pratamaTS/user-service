package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
)

type StockTransferRepository interface {
	Create(req *dao.StockTransfer) (dao.StockTransfer, error)
	Detail(uuid string) (dao.StockTransfer, error)
	List(req *dto.APIRequest[dto.FilterRequest]) ([]dao.StockTransfer, error)
	Receive(uuid string, receivedBy string) (dao.StockTransfer, error)
}

type StockTransferRepositoryImpl struct {
	transferCol *mongo.Collection
	stockCol    *mongo.Collection
	productCol  *mongo.Collection
	mongoClient *mongo.Client
}

func StockTransferRepositoryInit(mongoClient *mongo.Client) *StockTransferRepositoryImpl {
	dbName := helpers.ProvideDBName()
	db := mongoClient.Database(dbName)

	return &StockTransferRepositoryImpl{
		transferCol: db.Collection("stock_transfers"),
		stockCol:    db.Collection("branch_stocks"),
		productCol:  db.Collection("products"),
		mongoClient: mongoClient,
	}
}

// helper: get product + conversion for unit
func (r *StockTransferRepositoryImpl) getConversionToBase(ctx context.Context, branchUUID, productUUID, unit string) (sku, name string, baseUnit string, conv int64, err error) {
	var p struct {
		UUID       string `bson:"uuid"`
		BranchUUID string `bson:"branch_uuid"`
		SKU        string `bson:"sku"`
		Name       string `bson:"name"`
		BaseUnit   string `bson:"base_unit"`
		Units      []struct {
			Name             string `bson:"name"`
			ConversionToBase int64  `bson:"conversion_to_base"`
		} `bson:"units"`
	}
	// product harus milik branch FROM (biar konsisten)
	if err = r.productCol.FindOne(ctx, bson.M{"uuid": productUUID, "branch_uuid": branchUUID}).Decode(&p); err != nil {
		return "", "", "", 0, err
	}

	sku, name, baseUnit = p.SKU, p.Name, p.BaseUnit
	// cari unit
	for _, u := range p.Units {
		if u.Name == unit {
			if u.ConversionToBase <= 0 {
				return sku, name, baseUnit, 0, errors.New("invalid conversion_to_base")
			}
			return sku, name, baseUnit, u.ConversionToBase, nil
		}
	}
	return sku, name, baseUnit, 0, errors.New("unit not found in product")
}

func (r *StockTransferRepositoryImpl) Create(req *dao.StockTransfer) (dao.StockTransfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	// generate uuid
	req.UUID = helpers.GenerateUUID()
	req.Status = dao.StockTransferInProgress
	req.UpdatedAt = now.Unix()
	req.UpdatedAtStr = nowStr
	req.CreatedAt = now
	req.CreatedAtStr = nowStr

	// enrich items with conversion + qty_base + sku/name
	outItems := make([]dao.StockTransferItem, 0, len(req.Items))
	for _, it := range req.Items {
		sku, name, _, conv, err := r.getConversionToBase(ctx, req.FromBranchUUID, it.ProductUUID, it.Unit)
		if err != nil {
			return dao.StockTransfer{}, err
		}
		qtyBase := it.Qty * conv
		outItems = append(outItems, dao.StockTransferItem{
			ProductUUID:      it.ProductUUID,
			SKU:              sku,
			Name:             name,
			Unit:             it.Unit,
			Qty:              it.Qty,
			ConversionToBase: conv,
			QtyBase:          qtyBase,
		})
	}
	req.Items = outItems

	if _, err := r.transferCol.InsertOne(ctx, req); err != nil {
		return dao.StockTransfer{}, err
	}

	var out dao.StockTransfer
	if err := r.transferCol.FindOne(ctx, bson.M{"uuid": req.UUID}).Decode(&out); err != nil {
		return dao.StockTransfer{}, err
	}
	return out, nil
}

func (r *StockTransferRepositoryImpl) Detail(uuid string) (dao.StockTransfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.StockTransfer
	err := r.transferCol.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	return result, err
}

func (r *StockTransferRepositoryImpl) List(req *dto.APIRequest[dto.FilterRequest]) ([]dao.StockTransfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"uuid": bson.M{"$regex": req.Search, "$options": "i"}},
			{"notes": bson.M{"$regex": req.Search, "$options": "i"}},
			{"items.name": bson.M{"$regex": req.Search, "$options": "i"}},
			{"items.sku": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	for k, v := range req.FilterBy.FilterBy {
		if k == "" || v == nil {
			continue
		}
		filter[k] = v
	}

	sort := bson.D{}
	for k, v := range req.SortBy.SortBy {
		switch tv := v.(type) {
		case string:
			if tv == "asc" || tv == "ASC" || tv == "1" {
				sort = append(sort, bson.E{Key: k, Value: 1})
			} else {
				sort = append(sort, bson.E{Key: k, Value: -1})
			}
		default:
			sort = append(sort, bson.E{Key: k, Value: -1})
		}
	}
	if len(sort) == 0 {
		sort = bson.D{{Key: "created_at", Value: -1}}
	}

	page := req.Pagination.Pagination.Page
	size := req.Pagination.Pagination.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	skip := int64((page - 1) * size)

	opts := options.Find().SetSort(sort).SetSkip(skip).SetLimit(int64(size))

	cur, err := r.transferCol.Find(ctx, filter, opts)
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

// Receive: set DONE + mutate stock from->to in one transaction
func (r *StockTransferRepositoryImpl) Receive(uuid string, receivedBy string) (dao.StockTransfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	sess, err := r.mongoClient.StartSession()
	if err != nil {
		return dao.StockTransfer{}, err
	}
	defer sess.EndSession(ctx)

	var result dao.StockTransfer

	_, err = sess.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		// lock-ish by checking status
		var trx dao.StockTransfer
		if err := r.transferCol.FindOne(sc, bson.M{"uuid": uuid}).Decode(&trx); err != nil {
			return nil, err
		}
		if trx.Status != dao.StockTransferInProgress {
			return nil, errors.New("transfer already processed")
		}

		// cek stok cukup di from_branch
		for _, it := range trx.Items {
			var st dao.BranchStock
			_ = r.stockCol.FindOne(sc, bson.M{
				"branch_uuid":  trx.FromBranchUUID,
				"product_uuid": it.ProductUUID,
			}).Decode(&st)

			available := st.QtyBase
			if available < it.QtyBase {
				return nil, errors.New("insufficient stock for product " + it.ProductUUID)
			}
		}

		// mutate stock
		for _, it := range trx.Items {
			// from_branch -qty
			_, err := r.stockCol.UpdateOne(sc,
				bson.M{"branch_uuid": trx.FromBranchUUID, "product_uuid": it.ProductUUID},
				bson.M{
					"$inc": bson.M{"qty_base": -it.QtyBase},
					"$setOnInsert": bson.M{
						"uuid": helpers.GenerateUUID(),
					},
				},
				options.Update().SetUpsert(true),
			)
			if err != nil {
				return nil, err
			}

			// to_branch +qty
			_, err = r.stockCol.UpdateOne(sc,
				bson.M{"branch_uuid": trx.ToBranchUUID, "product_uuid": it.ProductUUID},
				bson.M{
					"$inc": bson.M{"qty_base": it.QtyBase},
					"$setOnInsert": bson.M{
						"uuid": helpers.GenerateUUID(),
					},
				},
				options.Update().SetUpsert(true),
			)
			if err != nil {
				return nil, err
			}
		}

		now := time.Now()
		nowStr := now.Format(time.RFC3339)

		// update transfer status
		_, err := r.transferCol.UpdateOne(sc,
			bson.M{"uuid": uuid, "status": dao.StockTransferInProgress},
			bson.M{
				"$set": bson.M{
					"status":          dao.StockTransferDone,
					"received_by":     receivedBy,
					"received_at":     now.Unix(),
					"received_at_str": nowStr,
					"updated_at":      now.Unix(),
					"updated_at_str":  nowStr,
				},
			},
		)
		if err != nil {
			return nil, err
		}

		if err := r.transferCol.FindOne(sc, bson.M{"uuid": uuid}).Decode(&result); err != nil {
			return nil, err
		}

		return true, nil
	})

	if err != nil {
		return dao.StockTransfer{}, err
	}
	return result, nil
}

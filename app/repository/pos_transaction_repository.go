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

type POSTransactionRepository interface {
	Insert(trx *dao.POSTransaction) (dao.POSTransaction, error)
	Detail(uuid string) (dao.POSTransaction, error)
	List(req *dto.FilterRequest) ([]dao.POSTransaction, error)
	UpdateStatus(uuid string, status string, extra bson.M) (dao.POSTransaction, error)
}

type POSTransactionRepositoryImpl struct {
	col *mongo.Collection
}

func POSTransactionRepositoryInit(mongoClient *mongo.Client) *POSTransactionRepositoryImpl {
	dbName := helpers.ProvideDBName()
	collection := mongoClient.Database(dbName).Collection("pos_transactions")
	return &POSTransactionRepositoryImpl{col: collection}
}

func (r *POSTransactionRepositoryImpl) Insert(trx *dao.POSTransaction) (dao.POSTransaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if trx.BranchUUID == "" {
		return dao.POSTransaction{}, errors.New("branch_uuid required")
	}
	if trx.PaymentMethod == "" {
		return dao.POSTransaction{}, errors.New("payment_method required")
	}
	if len(trx.Items) == 0 {
		return dao.POSTransaction{}, errors.New("items required")
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	if trx.UUID == "" {
		trx.UUID = helpers.GenerateUUID()
	}
	if trx.CreatedAt.IsZero() {
		trx.CreatedAt = now
	}
	if trx.CreatedAtStr == "" {
		trx.CreatedAtStr = nowStr
	}
	trx.UpdatedAt = now.Unix()
	trx.UpdatedAtStr = nowStr

	if _, err := r.col.InsertOne(ctx, trx); err != nil {
		return dao.POSTransaction{}, err
	}

	var out dao.POSTransaction
	if err := r.col.FindOne(ctx, bson.M{"uuid": trx.UUID}).Decode(&out); err != nil {
		return dao.POSTransaction{}, err
	}
	return out, nil
}

func (r *POSTransactionRepositoryImpl) Detail(uuid string) (dao.POSTransaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var out dao.POSTransaction
	err := r.col.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&out)
	return out, err
}

func (r *POSTransactionRepositoryImpl) List(req *dto.FilterRequest) ([]dao.POSTransaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{}

	// search
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"receipt_no": bson.M{"$regex": req.Search, "$options": "i"}},
			{"payment_method": bson.M{"$regex": req.Search, "$options": "i"}},
			{"items.name": bson.M{"$regex": req.Search, "$options": "i"}},
			{"items.sku": bson.M{"$regex": req.Search, "$options": "i"}},
			{"items.barcode": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	// filter_by (langsung)
	for k, v := range req.FilterBy {
		if k == "" || v == nil {
			continue
		}
		filter[k] = v
	}

	// sort_by
	sort := bson.D{}
	for k, v := range req.SortBy {
		switch tv := v.(type) {
		case string:
			if tv == "asc" || tv == "ASC" || tv == "1" {
				sort = append(sort, bson.E{Key: k, Value: 1})
			} else {
				sort = append(sort, bson.E{Key: k, Value: -1})
			}
		case float64:
			if int(tv) >= 0 {
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

	// pagination (punya kamu)
	page := req.Pagination.Page
	size := req.Pagination.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	skip := int64((page - 1) * size)

	opts := options.Find().SetSort(sort).SetSkip(skip).SetLimit(int64(size))

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []dao.POSTransaction
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *POSTransactionRepositoryImpl) UpdateStatus(uuid string, status string, extra bson.M) (dao.POSTransaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	set := bson.M{
		"status":         status,
		"updated_at":     now.Unix(),
		"updated_at_str": nowStr,
	}
	for k, v := range extra {
		set[k] = v
	}

	if _, err := r.col.UpdateOne(ctx, bson.M{"uuid": uuid}, bson.M{"$set": set}); err != nil {
		return dao.POSTransaction{}, err
	}

	var out dao.POSTransaction
	if err := r.col.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&out); err != nil {
		return dao.POSTransaction{}, err
	}
	return out, nil
}

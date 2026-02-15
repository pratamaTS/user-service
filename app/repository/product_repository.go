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

type ProductRepository interface {
	SaveProduct(data *dao.Product) (dao.Product, error)
	BulkUpsertProducts(list []dao.Product) (int, int, []error)
	DetailProduct(uuid string) (dao.Product, error)
	ListProduct(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Product, error)
	DeleteProduct(uuid string) error
}

type ProductRepositoryImpl struct {
	productCollection *mongo.Collection
}

func ProductRepositoryInit(mongoClient *mongo.Client) *ProductRepositoryImpl {
	dbName := helpers.ProvideDBName()
	collection := mongoClient.Database(dbName).Collection("products")
	return &ProductRepositoryImpl{
		productCollection: collection,
	}
}

func (r *ProductRepositoryImpl) SaveProduct(data *dao.Product) (dao.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if data.BranchUUID == "" {
		return dao.Product{}, errors.New("branch_uuid required")
	}
	if data.Name == "" {
		return dao.Product{}, errors.New("name required")
	}
	if data.BaseUnit == "" {
		return dao.Product{}, errors.New("base_unit required")
	}

	filter := bson.M{}
	switch {
	case data.UUID != "":
		filter["uuid"] = data.UUID
	case data.SKU != "":
		filter["branch_uuid"] = data.BranchUUID
		filter["sku"] = data.SKU
	default:
		filter["branch_uuid"] = data.BranchUUID
		filter["name"] = data.Name
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	newUUID := data.UUID
	if newUUID == "" {
		newUUID = helpers.GenerateUUID()
	}

	update := bson.M{
		"$set": bson.M{
			"branch_uuid":    data.BranchUUID,
			"image":          data.Image,
			"sku":            data.SKU,
			"barcode":        data.Barcode,
			"name":           data.Name,
			"description":    data.Description,
			"base_unit":      data.BaseUnit,
			"units":          data.Units,
			"cost":           data.Cost,
			"price":          data.Price,
			"is_active":      data.IsActive,
			"created_by":     data.CreatedBy,
			"updated_at":     now.Unix(),
			"updated_at_str": nowStr,
		},
		"$setOnInsert": bson.M{
			"uuid": newUUID,
			"created_at": func() time.Time {
				if data.CreatedAt.IsZero() {
					return now
				}
				return data.CreatedAt
			}(),
			"created_at_str": func() string {
				if data.CreatedAtStr == "" {
					return nowStr
				}
				return data.CreatedAtStr
			}(),
		},
	}

	opts := options.Update().SetUpsert(true)
	if _, err := r.productCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return dao.Product{}, err
	}

	var out dao.Product
	if err := r.productCollection.FindOne(ctx, filter).Decode(&out); err != nil {
		return dao.Product{}, err
	}
	return out, nil
}

// Bulk: per-row upsert, return (successCount, failCount, errors)
func (r *ProductRepositoryImpl) BulkUpsertProducts(list []dao.Product) (int, int, []error) {
	ok := 0
	fail := 0
	var errs []error

	for i := range list {
		_, err := r.SaveProduct(&list[i])
		if err != nil {
			fail++
			errs = append(errs, err)
			continue
		}
		ok++
	}
	return ok, fail, errs
}

func (r *ProductRepositoryImpl) DetailProduct(uuid string) (dao.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.Product
	err := r.productCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	return result, err
}

func (r *ProductRepositoryImpl) ListProduct(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": req.Search, "$options": "i"}},
			{"sku": bson.M{"$regex": req.Search, "$options": "i"}},
			{"barcode": bson.M{"$regex": req.Search, "$options": "i"}},
			{"base_unit": bson.M{"$regex": req.Search, "$options": "i"}},
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
		case float64:
			if int(tv) >= 0 {
				sort = append(sort, bson.E{Key: k, Value: 1})
			} else {
				sort = append(sort, bson.E{Key: k, Value: -1})
			}
		default:
			sort = append(sort, bson.E{Key: k, Value: 1})
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

	cur, err := r.productCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var listOut []dao.Product
	if err := cur.All(ctx, &listOut); err != nil {
		return nil, err
	}
	return listOut, nil
}

func (r *ProductRepositoryImpl) DeleteProduct(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.productCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

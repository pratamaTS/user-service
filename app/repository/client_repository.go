package repository

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
)

type ClientRepository interface {
	SaveClient(data *dao.Client) (dao.Client, error)
	DetailClient(uuid string) (dao.Client, error)
	ListClient(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Client, error)
	DeleteClient(uuid string) error
}

type ClientRepositoryImpl struct {
	clientCollection *mongo.Collection
}

func (u *ClientRepositoryImpl) SaveClient(data *dao.Client) (dao.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	filter := bson.M{
		"uuid": data.UUID,
	}

	data.UUID = helpers.GenerateUUID()
	data.UpdatedAt = now.Unix()
	data.UpdatedAtStr = nowStr

	update := bson.M{
		"$set": bson.M{
			"company_uuid":   data.CompanyUUID,
			"logo":           data.Logo,
			"name":           data.Name,
			"host":           data.Host,
			"website_url":    data.WebisteUrl,
			"phone_number":   data.PhoneNumber,
			"is_active":      data.IsActive,
			"updated_at":     data.UpdatedAt,
			"updated_at_str": data.UpdatedAtStr,
		},
		"$setOnInsert": bson.M{
			"uuid": data.UUID,
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
			"created_by": data.CreatedBy,
		},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var out dao.Client
	if err := u.clientCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out); err != nil {
		return dao.Client{}, err
	}

	return out, nil
}

func (u *ClientRepositoryImpl) DetailClient(uuid string) (dao.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.Client
	log.Println("DetailClient uuid:", uuid)
	err := u.clientCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	log.Println("DetailClient result:", result)
	return result, err
}

func (u *ClientRepositoryImpl) ListClient(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": req.Search, "$options": "i"}},
			{"url": bson.M{"$regex": req.Search, "$options": "i"}},
			{"phone_number": bson.M{"$regex": req.Search, "$options": "i"}},
			{"company_uuid": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}
	for key, val := range req.FilterBy.FilterBy {
		if key == "" || val == nil {
			continue
		}
		filter[key] = val
	}

	sort := bson.D{}
	for key, val := range req.SortBy.SortBy {
		switch v := val.(type) {
		case string:
			if v == "asc" || v == "ASC" || v == "1" {
				sort = append(sort, bson.E{Key: key, Value: 1})
			} else {
				sort = append(sort, bson.E{Key: key, Value: -1})
			}
		case float64:
			if int(v) >= 0 {
				sort = append(sort, bson.E{Key: key, Value: 1})
			} else {
				sort = append(sort, bson.E{Key: key, Value: -1})
			}
		default:
			sort = append(sort, bson.E{Key: key, Value: 1})
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

	opts := options.Find().
		SetSort(sort).
		SetSkip(skip).
		SetLimit(int64(size))

	cur, err := u.clientCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var clients []dao.Client
	if err := cur.All(ctx, &clients); err != nil {
		return nil, err
	}
	return clients, nil
}

func (u *ClientRepositoryImpl) DeleteClient(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := u.clientCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

func ClientRepositoryInit(mongoClient *mongo.Client) *ClientRepositoryImpl {
	dbName := helpers.ProvideDBName()
	clientCollection := mongoClient.Database(dbName).Collection("clients")
	return &ClientRepositoryImpl{
		clientCollection: clientCollection,
	}
}

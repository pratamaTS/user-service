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

type ClientBranchRepository interface {
	SaveClientBranch(data *dao.ClientBranch) (dao.ClientBranch, error)
	DetailClientBranch(uuid string) (dao.ClientBranch, error)
	ListClientBranch(req *dto.FilterRequest) ([]dao.ClientBranch, error)
	DeleteClientBranch(uuid string) error
}

type ClientBranchRepositoryImpl struct {
	clientBranchCollection *mongo.Collection
}

func (u *ClientBranchRepositoryImpl) SaveClientBranch(data *dao.ClientBranch) (dao.ClientBranch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	switch {
	case data.UUID != "":
		filter["uuid"] = data.UUID
	case data.ClientUUID != "" && data.Name != "":
		filter["client_uuid"] = data.ClientUUID
		filter["name"] = data.Name
	default:
		return dao.ClientBranch{}, errors.New("upsert requires uuid, id, or (client_uuid + name)")
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	data.UpdatedAt = now.Unix()
	data.UpdatedAtStr = nowStr
	data.UUID = helpers.GenerateUUID()

	update := bson.M{
		"$set": bson.M{
			"client_uuid":    data.ClientUUID,
			"name":           data.Name,
			"address":        data.Address,
			"longitude":      data.Longitude,
			"latitude":       data.Latitude,
			"phone_number":   data.PhoneNumber,
			"total_staff":    data.TotalStaff,
			"max_radius":     data.MaxRadius,
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
		},
	}

	opts := options.Update().SetUpsert(true)

	if _, err := u.clientBranchCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return dao.ClientBranch{}, err
	}

	var out dao.ClientBranch
	if err := u.clientBranchCollection.FindOne(ctx, filter).Decode(&out); err != nil {
		return dao.ClientBranch{}, err
	}
	return out, nil
}

func (u *ClientBranchRepositoryImpl) DetailClientBranch(uuid string) (dao.ClientBranch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.ClientBranch
	err := u.clientBranchCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	return result, err
}

func (u *ClientBranchRepositoryImpl) ListClientBranch(req *dto.FilterRequest) ([]dao.ClientBranch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"company_uuid": bson.M{"$regex": req.Search, "$options": "i"}},
			{"role_uuid": bson.M{"$regex": req.Search, "$options": "i"}},
			{"user_uuid": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}
	for k, v := range req.FilterBy {
		if k == "" || v == nil {
			continue
		}
		filter[k] = v
	}

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
			sort = append(sort, bson.E{Key: k, Value: 1})
		}
	}
	if len(sort) == 0 {
		sort = bson.D{{Key: "created_at", Value: -1}}
	}

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

	cur, err := u.clientBranchCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []dao.ClientBranch
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (u *ClientBranchRepositoryImpl) DeleteClientBranch(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := u.clientBranchCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

func ClientBranchRepositoryInit(mongoClient *mongo.Client) *ClientBranchRepositoryImpl {
	dbName := helpers.ProvideDBName()
	collection := mongoClient.Database(dbName).Collection("client_branches")
	return &ClientBranchRepositoryImpl{
		clientBranchCollection: collection,
	}
}

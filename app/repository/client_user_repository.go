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

type ClientUserRepository interface {
	SaveClientUser(data *dao.ClientUser) (dao.ClientUser, error)
	DetailClientUser(uuid string) (dao.ClientUser, error)
	ListClientUser(req *dto.FilterRequest) ([]dao.ClientUser, error)
	DeleteClientUser(uuid string) error
}

type ClientUserRepositoryImpl struct {
	clientUserCollection *mongo.Collection
}

func (u *ClientUserRepositoryImpl) SaveClientUser(data *dao.ClientUser) (dao.ClientUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	switch {
	case data.UUID != "":
		filter["uuid"] = data.UUID
	case data.ClientUUID != "" && data.UserUUID != "":
		filter["company_uuid"] = data.ClientUUID
		filter["$or"] = []bson.M{
			{"user_uuid": data.UserUUID},
		}
	default:
		return dao.ClientUser{}, errors.New("upsert requires uuid, id, or (client_uuid + user_uuid)")
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	data.UpdatedAt = now.Unix()
	data.UpdatedAtStr = nowStr
	data.UUID = helpers.GenerateUUID()
	update := bson.M{
		"$set": bson.M{
			"client_uuid":    data.ClientUUID,
			"branch_uuid":    data.BranchUUID,
			"role_uuid":      data.RoleUUID,
			"user_uuid":      data.UserUUID,
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

	if _, err := u.clientUserCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return dao.ClientUser{}, err
	}

	var out dao.ClientUser
	if err := u.clientUserCollection.FindOne(ctx, filter).Decode(&out); err != nil {
		return dao.ClientUser{}, err
	}
	return out, nil
}

func (u *ClientUserRepositoryImpl) DetailClientUser(uuid string) (dao.ClientUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.ClientUser
	err := u.clientUserCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	return result, err
}

func (u *ClientUserRepositoryImpl) ListClientUser(req *dto.FilterRequest) ([]dao.ClientUser, error) {
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

	cur, err := u.clientUserCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []dao.ClientUser
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (u *ClientUserRepositoryImpl) DeleteClientUser(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := u.clientUserCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

func ClientUserRepositoryInit(mongoClient *mongo.Client) *ClientUserRepositoryImpl {
	dbName := helpers.ProvideDBName()
	collection := mongoClient.Database(dbName).Collection("client_users")
	return &ClientUserRepositoryImpl{
		clientUserCollection: collection,
	}
}

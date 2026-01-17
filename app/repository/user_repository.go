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
)

type UserRepository interface {
	SaveUser(data *dao.User) (dao.User, error)
	DetailUser(uuid string) (dao.User, error)
	ListUser(req *dto.APIRequest[dto.FilterRequest]) ([]dao.User, error)
	DeleteUser(uuid string) error
}

type UserRepositoryImpl struct {
	userCollection *mongo.Collection
}

func (u *UserRepositoryImpl) SaveUser(data *dao.User) (dao.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	switch {
	case data.UUID != "":
		filter["uuid"] = data.UUID
	case data.Email != "":
		filter["email"] = data.Email
	default:
		return dao.User{}, errors.New("upsert requires one of uuid/id/email")
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	data.UpdatedAt = now.Unix()
	data.UpdatedAtStr = nowStr

	update := bson.M{
		"$set": bson.M{
			"image":          data.Image,
			"username":       data.Username,
			"name":           data.Name,
			"email":          data.Email,
			"password":       data.Password,
			"phone_number":   data.PhoneNumber,
			"address":        data.Address,
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

	if _, err := u.userCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return dao.User{}, err
	}

	var out dao.User
	if err := u.userCollection.FindOne(ctx, filter).Decode(&out); err != nil {
		return dao.User{}, err
	}
	return out, nil
}

func (u *UserRepositoryImpl) DetailUser(uuid string) (dao.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var result dao.User
	err := u.userCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	return result, err
}

func (u *UserRepositoryImpl) ListUser(req *dto.APIRequest[dto.FilterRequest]) ([]dao.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"username": bson.M{"$regex": req.Search, "$options": "i"}},
			{"name": bson.M{"$regex": req.Search, "$options": "i"}},
			{"email": bson.M{"$regex": req.Search, "$options": "i"}},
			{"phone_number": bson.M{"$regex": req.Search, "$options": "i"}},
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

	cur, err := u.userCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var users []dao.User
	if err := cur.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UserRepositoryImpl) DeleteUser(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := u.userCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

func UserRepositoryInit(mongoClient *mongo.Client) *UserRepositoryImpl {
	userCollection := mongoClient.Database("db_portal_general").Collection("cl_users")
	return &UserRepositoryImpl{
		userCollection: userCollection,
	}
}

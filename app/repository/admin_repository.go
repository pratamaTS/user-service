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

type AdminRepository interface {
	SaveAdmin(data *dao.Admin) (dao.Admin, error)
	DetailAdmin(uuid string) (dao.Admin, error)
	ListAdmin(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Admin, error)
	DeleteAdmin(uuid string) error
}

type AdminRepositoryImpl struct {
	adminCollection *mongo.Collection
}

func (u *AdminRepositoryImpl) SaveAdmin(data *dao.Admin) (dao.Admin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	switch {
	case data.UUID != "":
		filter["uuid"] = data.UUID
	case data.Email != "":
		filter["email"] = data.Email
	default:
		return dao.Admin{}, errors.New("upsert requires one of uuid/id/email")
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

	if _, err := u.adminCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return dao.Admin{}, err
	}

	var out dao.Admin
	if err := u.adminCollection.FindOne(ctx, filter).Decode(&out); err != nil {
		return dao.Admin{}, err
	}
	return out, nil
}

func (u *AdminRepositoryImpl) DetailAdmin(uuid string) (dao.Admin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.Admin
	err := u.adminCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	return result, err
}

func (u *AdminRepositoryImpl) ListAdmin(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Admin, error) {
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

	cur, err := u.adminCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var admins []dao.Admin
	if err := cur.All(ctx, &admins); err != nil {
		return nil, err
	}
	return admins, nil
}

func (u *AdminRepositoryImpl) DeleteAdmin(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := u.adminCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

func AdminRepositoryInit(mongoClient *mongo.Client) *AdminRepositoryImpl {
	dbName := helpers.ProvideDBName()
	adminCollection := mongoClient.Database(dbName).Collection("admins")
	return &AdminRepositoryImpl{
		adminCollection: adminCollection,
	}
}

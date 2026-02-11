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

type ParentMenuRepository interface {
	SaveParentMenu(data *dao.ParentMenu) (dao.ParentMenu, error)
	DetailParentMenu(uuid string) (dao.ParentMenu, error)
	ListParentMenu(req *dto.FilterRequest) ([]dao.ParentMenu, error)
	DeleteParentMenu(uuid string) error
}

type ParentMenuRepositoryImpl struct {
	parentMenuCollection *mongo.Collection
}

func (u *ParentMenuRepositoryImpl) SaveParentMenu(data *dao.ParentMenu) (dao.ParentMenu, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	switch {
	case data.UUID != "":
		filter["uuid"] = data.UUID
	case data.Title != "":
		filter["title"] = data.Title
	default:
		return dao.ParentMenu{}, errors.New("upsert requires one of uuid/id/title")
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	data.UUID = helpers.GenerateUUID()
	data.UpdatedAt = now.Unix()
	data.UpdatedAtStr = nowStr

	update := bson.M{
		"$set": bson.M{
			"icon":           data.Icon,
			"title":          data.Title,
			"description":    data.Description,
			"sort":           data.Sort,
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

	if _, err := u.parentMenuCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return dao.ParentMenu{}, err
	}

	var out dao.ParentMenu
	if err := u.parentMenuCollection.FindOne(ctx, filter).Decode(&out); err != nil {
		return dao.ParentMenu{}, err
	}
	return out, nil
}

func (u *ParentMenuRepositoryImpl) DetailParentMenu(uuid string) (dao.ParentMenu, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.ParentMenu
	err := u.parentMenuCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	return result, err
}

func (u *ParentMenuRepositoryImpl) ListParentMenu(req *dto.FilterRequest) ([]dao.ParentMenu, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": req.Search, "$options": "i"}},
			{"description": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}
	for key, val := range req.FilterBy {
		if key == "" || val == nil {
			continue
		}
		filter[key] = val
	}

	sort := bson.D{}
	for key, val := range req.SortBy {
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
		sort = bson.D{{Key: "sort", Value: 1}} // Default: order by sort ascending
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

	opts := options.Find().
		SetSort(sort).
		SetSkip(skip).
		SetLimit(int64(size))

	cur, err := u.parentMenuCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var menus []dao.ParentMenu
	if err := cur.All(ctx, &menus); err != nil {
		return nil, err
	}
	return menus, nil
}

func (u *ParentMenuRepositoryImpl) DeleteParentMenu(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := u.parentMenuCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

func ParentMenuRepositoryInit(mongoClient *mongo.Client) *ParentMenuRepositoryImpl {
	dbName := helpers.ProvideDBName()
	parentMenuCollection := mongoClient.Database(dbName).Collection("parent_menus")
	return &ParentMenuRepositoryImpl{
		parentMenuCollection: parentMenuCollection,
	}
}

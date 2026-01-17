package repository

import (
	"context"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
)

type RoleRepository interface {
	SaveRole(data *dao.Role) (dao.Role, error)
	DetailRole(uuid string) (dao.Role, error)
	GetRoleByValue(value, subject string) (dao.Role, error)
	ListRole(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Role, error)
	DeleteRole(uuid string) error
}

type RoleRepositoryImpl struct {
	roleCollection *mongo.Collection
}

func (u *RoleRepositoryImpl) SaveRole(data *dao.Role) (dao.Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}

	filter["name"] = data.Name
	filter["value"] = data.Value
	filter["is_use_by"] = data.IsUseBy

	log.Println("Role upsert filter:", filter)

	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	data.UUID = helpers.GenerateUUID()
	data.UpdatedAt = now.Unix()
	data.UpdatedAtStr = nowStr

	update := bson.M{
		"$set": bson.M{
			"name":           data.Name,
			"value":          data.Value,
			"is_use_by":      data.IsUseBy,
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

	if _, err := u.roleCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return dao.Role{}, err
	}

	var out dao.Role
	if err := u.roleCollection.FindOne(ctx, filter).Decode(&out); err != nil {
		return dao.Role{}, err
	}
	return out, nil
}

func (u *RoleRepositoryImpl) DetailRole(uuid string) (dao.Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.Role
	err := u.roleCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	return result, err
}

func (u *RoleRepositoryImpl) GetRoleByValue(value, subject string) (dao.Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.Role
	err := u.roleCollection.FindOne(ctx, bson.M{"value": value, "is_use_by": strings.ToLower(subject)}).Decode(&result)
	return result, err
}

func (u *RoleRepositoryImpl) ListRole(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": req.Search, "$options": "i"}},
			{"value": bson.M{"$regex": req.Search, "$options": "i"}},
			{"is_use_by": bson.M{"$regex": req.Search, "$options": "i"}},
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

	cur, err := u.roleCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var roles []dao.Role
	if err := cur.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (u *RoleRepositoryImpl) DeleteRole(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := u.roleCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

func RoleRepositoryInit(mongoClient *mongo.Client) *RoleRepositoryImpl {
	roleCollection := mongoClient.Database("db_portal_general").Collection("cl_roles")
	return &RoleRepositoryImpl{
		roleCollection: roleCollection,
	}
}

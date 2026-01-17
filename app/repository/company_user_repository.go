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

type CompanyUserRepository interface {
	SaveCompanyUser(data *dao.CompanyUser) (dao.CompanyUser, error)
	DetailCompanyUser(uuid string) (dao.CompanyUser, error)
	ListCompanyUser(req *dto.APIRequest[dto.FilterRequest]) ([]dao.CompanyUser, error)
	DeleteCompanyUser(uuid string) error
}

type CompanyUserRepositoryImpl struct {
	companyUserCollection *mongo.Collection
}

func (u *CompanyUserRepositoryImpl) SaveCompanyUser(data *dao.CompanyUser) (dao.CompanyUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	switch {
	case data.UUID != "":
		filter["uuid"] = data.UUID
	case data.CompanyUUID != "" && data.AdminUUID != "":
		filter["company_uuid"] = data.CompanyUUID
		filter["user_uuid"] = data.AdminUUID
	default:
		return dao.CompanyUser{}, errors.New("upsert requires uuid, id, or (company_uuid + admin_uuid)")
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	data.UpdatedAt = now.Unix()
	data.UpdatedAtStr = nowStr

	update := bson.M{
		"$set": bson.M{
			"company_uuid":   data.CompanyUUID,
			"role_uuid":      data.RoleUUID,
			"user_uuid":      data.AdminUUID,
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

	if _, err := u.companyUserCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return dao.CompanyUser{}, err
	}

	var out dao.CompanyUser
	if err := u.companyUserCollection.FindOne(ctx, filter).Decode(&out); err != nil {
		return dao.CompanyUser{}, err
	}
	return out, nil
}

func (u *CompanyUserRepositoryImpl) DetailCompanyUser(uuid string) (dao.CompanyUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.CompanyUser
	err := u.companyUserCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	return result, err
}

func (u *CompanyUserRepositoryImpl) ListCompanyUser(req *dto.APIRequest[dto.FilterRequest]) ([]dao.CompanyUser, error) {
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

	cur, err := u.companyUserCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []dao.CompanyUser
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (u *CompanyUserRepositoryImpl) DeleteCompanyUser(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := u.companyUserCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

func CompanyUserRepositoryInit(mongoClient *mongo.Client) *CompanyUserRepositoryImpl {
	collection := mongoClient.Database("db_portal_general").Collection("cl_company_users")
	return &CompanyUserRepositoryImpl{
		companyUserCollection: collection,
	}
}

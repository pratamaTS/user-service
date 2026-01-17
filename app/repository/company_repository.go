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

type CompanyRepository interface {
	SaveCompany(data *dao.Company) (dao.Company, error)
	DetailCompany(uuid string) (dao.Company, error)
	ListCompany(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Company, error)
	DeleteCompany(uuid string) error
}

type CompanyRepositoryImpl struct {
	companyCollection *mongo.Collection
}

func (u *CompanyRepositoryImpl) SaveCompany(data *dao.Company) (dao.Company, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if data.UUID == "" {
		data.UUID = helpers.GenerateUUID()
	}

	filter := bson.M{
		"uuid": data.UUID,
	}

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	data.UpdatedAt = now.Unix()
	data.UpdatedAtStr = nowStr

	update := bson.M{
		"$set": bson.M{
			"logo":           data.Logo,
			"name":           data.Name,
			"host":           data.Host,
			"website_url":    data.WebisteUrl,
			"phone_number":   data.PhoneNumber,
			"address":        data.Address,
			"website":        data.Website,
			"nib":            data.NIB,
			"npwp":           data.NPWP,
			"email_company":  data.EmailCompany,
			"email_notif":    data.EmailNotif,
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

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var out dao.Company
	if err := u.companyCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out); err != nil {
		return dao.Company{}, err
	}

	return out, nil
}

func (u *CompanyRepositoryImpl) DetailCompany(uuid string) (dao.Company, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.Company
	err := u.companyCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	return result, err
}

func (u *CompanyRepositoryImpl) ListCompany(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Company, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": req.Search, "$options": "i"}},
			{"url": bson.M{"$regex": req.Search, "$options": "i"}},
			{"email_company": bson.M{"$regex": req.Search, "$options": "i"}},
			{"email_notif": bson.M{"$regex": req.Search, "$options": "i"}},
			{"nib": bson.M{"$regex": req.Search, "$options": "i"}},
			{"npwp": bson.M{"$regex": req.Search, "$options": "i"}},
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

	log.Print("Filter company: ", filter)

	// 4️⃣ Execute query
	cur, err := u.companyCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var companies []dao.Company
	if err := cur.All(ctx, &companies); err != nil {
		return nil, err
	}
	return companies, nil
}

func (u *CompanyRepositoryImpl) DeleteCompany(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := u.companyCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

func CompanyRepositoryInit(mongoClient *mongo.Client) *CompanyRepositoryImpl {
	companyCollection := mongoClient.Database("db_portal_general").Collection("cl_companies")
	return &CompanyRepositoryImpl{
		companyCollection: companyCollection,
	}
}

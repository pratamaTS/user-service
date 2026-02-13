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

type SubscriptionRepository interface {
	Upsert(data *dao.Subscription) (dao.Subscription, error)
	Detail(uuid string) (dao.Subscription, error)
	List(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Subscription, error)
	Delete(uuid string) error
}

type SubscriptionRepositoryImpl struct {
	subscriptionCollection *mongo.Collection
}

func (r *SubscriptionRepositoryImpl) Upsert(data *dao.Subscription) (dao.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	if data.UUID == "" {
		data.UUID = helpers.GenerateUUID()
	}

	if data.Name == "" {
		return dao.Subscription{}, errors.New("name required")
	}
	if data.Type == "" {
		return dao.Subscription{}, errors.New("type required")
	}

	switch data.Type {
	case dao.SubscriptionTrial:
		data.BillingPeriod = dao.BillingWeekly
		data.DurationDays = 7
		data.Price = 0
		data.IsFeatureFull = false
	case dao.SubscriptionBusiness:
		if data.BillingPeriod != dao.BillingMonthly && data.BillingPeriod != dao.BillingYearly {
			data.BillingPeriod = dao.BillingMonthly
		}
		if data.DurationDays <= 0 {
			if data.BillingPeriod == dao.BillingYearly {
				data.DurationDays = 365
			} else {
				data.DurationDays = 30
			}
		}
		data.IsFeatureFull = false
	case dao.SubscriptionCompany:
		if data.BillingPeriod != dao.BillingMonthly && data.BillingPeriod != dao.BillingYearly {
			data.BillingPeriod = dao.BillingMonthly
		}
		if data.DurationDays <= 0 {
			if data.BillingPeriod == dao.BillingYearly {
				data.DurationDays = 365
			} else {
				data.DurationDays = 30
			}
		}
		data.IsFeatureFull = true
	default:
		return dao.Subscription{}, errors.New("invalid subscription type")
	}

	data.UpdatedAt = now.Unix()
	data.UpdatedAtStr = nowStr

	filter := bson.M{"uuid": data.UUID}

	update := bson.M{
		"$set": bson.M{
			"uuid":            data.UUID,
			"name":            data.Name,
			"type":            data.Type,
			"billing_period":  data.BillingPeriod,
			"duration_days":   data.DurationDays,
			"price":           data.Price,
			"is_feature_full": data.IsFeatureFull,
			"max_user":        data.MaxUser,
			"max_project":     data.MaxProject,
			"max_storage_gb":  data.MaxStorageGB,
			"is_active":       data.IsActive,
			"created_by":      data.CreatedBy,
			"updated_at":      data.UpdatedAt,
			"updated_at_str":  data.UpdatedAtStr,
		},
		"$setOnInsert": bson.M{
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

	var out dao.Subscription
	if err := r.subscriptionCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out); err != nil {
		return dao.Subscription{}, err
	}
	return out, nil
}

func (r *SubscriptionRepositoryImpl) Detail(uuid string) (dao.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var res dao.Subscription
	err := r.subscriptionCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&res)
	return res, err
}

func (r *SubscriptionRepositoryImpl) List(req *dto.APIRequest[dto.FilterRequest]) ([]dao.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": req.Search, "$options": "i"}},
			{"type": bson.M{"$regex": req.Search, "$options": "i"}},
			{"billing_period": bson.M{"$regex": req.Search, "$options": "i"}},
			{"created_by": bson.M{"$regex": req.Search, "$options": "i"}},
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

	opts := options.Find().SetSort(sort).SetSkip(skip).SetLimit(int64(size))

	cur, err := r.subscriptionCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []dao.Subscription
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *SubscriptionRepositoryImpl) Delete(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.subscriptionCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

func SubscriptionRepositoryInit(mongoClient *mongo.Client) *SubscriptionRepositoryImpl {
	dbName := helpers.ProvideDBName()
	subscriptionCollection := mongoClient.Database(dbName).Collection("subscriptions")
	return &SubscriptionRepositoryImpl{
		subscriptionCollection: subscriptionCollection,
	}
}

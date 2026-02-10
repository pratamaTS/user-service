package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
)

type ClientSubscriptionRepository interface {
	ActivateFromMaster(clientUUID, subscriptionUUID, createdBy string) (dao.ClientSubscription, error)
	GetActiveByClientUUID(clientUUID string) (*dao.ClientSubscription, error)
	ListClientBySubscriptionUUID(req *dto.FilterRequest) ([]dao.ClientSubscriptionJoined, error)
	IsClientAllowed(clientUUID string) (bool, *dao.ClientSubscription, error)
}

type ClientSubscriptionRepositoryImpl struct {
	col          *mongo.Collection
	subMasterCol *mongo.Collection
}

func (r *ClientSubscriptionRepositoryImpl) ActivateFromMaster(clientUUID, subscriptionUUID, createdBy string) (dao.ClientSubscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if clientUUID == "" || subscriptionUUID == "" {
		return dao.ClientSubscription{}, errors.New("client_uuid & subscription_uuid required")
	}

	var master dao.Subscription
	if err := r.subMasterCol.FindOne(ctx, bson.M{"uuid": subscriptionUUID, "is_active": true}).Decode(&master); err != nil {
		return dao.ClientSubscription{}, errors.New("master subscription not found / inactive")
	}

	now := time.Now()
	startAt := now
	expiredAt := startAt.AddDate(0, 0, master.DurationDays)
	isActive := expiredAt.After(now)

	filter := bson.M{"client_uuid": clientUUID}

	update := bson.M{
		"$set": bson.M{
			"client_uuid":       clientUUID,
			"subscription_uuid": subscriptionUUID,
			"start_at":          startAt,
			"expired_at":        expiredAt,
			"is_active":         isActive,
			"created_by":        createdBy,
			"updated_at":        now.Unix(),
			"updated_at_str":    now.Format(time.RFC3339),
		},
		"$setOnInsert": bson.M{
			"uuid":           helpers.GenerateUUID(),
			"created_at":     now,
			"created_at_str": now.Format(time.RFC3339),
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var out dao.ClientSubscription
	if err := r.col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out); err != nil {
		return dao.ClientSubscription{}, err
	}

	return out, nil
}

func (r *ClientSubscriptionRepositoryImpl) GetActiveByClientUUID(clientUUID string) (*dao.ClientSubscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if clientUUID == "" {
		return nil, errors.New("client_uuid required")
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var sub dao.ClientSubscription
	err := r.col.FindOne(ctx, bson.M{"client_uuid": clientUUID}, opts).Decode(&sub)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if !sub.ExpiredAt.IsZero() && now.After(sub.ExpiredAt) && sub.IsActive {
		_, _ = r.col.UpdateOne(ctx,
			bson.M{"uuid": sub.UUID},
			bson.M{"$set": bson.M{
				"is_active":      false,
				"updated_at":     now.Unix(),
				"updated_at_str": now.Format(time.RFC3339),
			}},
		)
		sub.IsActive = false
	}

	return &sub, nil
}

func (r *ClientSubscriptionRepositoryImpl) ListClientBySubscriptionUUID(req *dto.FilterRequest) ([]dao.ClientSubscriptionJoined, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ambil subscription_uuid dari filter_by (karena payload kamu "sama kayak list/fetch")
	subUUIDAny := req.FilterBy["subscription_uuid"]
	subscriptionUUID, _ := subUUIDAny.(string)
	if subscriptionUUID == "" {
		return nil, errors.New("subscription_uuid required (put it in filter_by.subscription_uuid)")
	}

	// =========================
	// 1) MATCH (FILTER)
	// =========================
	filter := bson.M{
		"subscription_uuid": subscriptionUUID,
	}

	// apply filter_by lainnya (kecuali subscription_uuid supaya ga dobel)
	for key, val := range req.FilterBy {
		if key == "" || val == nil || key == "subscription_uuid" {
			continue
		}

		// support filter khusus untuk client.* dengan prefix "client."
		// contoh: filter_by: { "client.is_active": true }
		filter[key] = val
	}

	// =========================
	// 2) SEARCH (regex i)
	// =========================
	// Search bisa kena field di client_subscriptions dan field client hasil lookup
	if req.Search != "" {
		reg := bson.M{"$regex": req.Search, "$options": "i"}
		filter["$or"] = []bson.M{
			// fields di client_subscriptions
			{"client_uuid": reg},
			{"created_by": reg},
			{"type": reg},
			{"billing_period": reg},

			// fields di client (hasil lookup)
			{"client.name": reg},
			{"client.url": reg},          // kalau field kamu "url"
			{"client.host": reg},         // kalau field kamu "host"
			{"client.phone_number": reg}, // kalau field kamu "phone_number"
		}
	}

	// =========================
	// 3) SORT (ikut template kamu)
	// =========================
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
		sort = bson.D{{Key: "created_at", Value: -1}}
	}

	// =========================
	// 4) PAGINATION (ikut template kamu)
	// =========================
	page := req.Pagination.Page
	size := req.Pagination.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	skip := int64((page - 1) * size)

	// =========================
	// 5) AGGREGATION PIPELINE
	// =========================
	pipeline := mongo.Pipeline{
		// match + search
		{{Key: "$match", Value: filter}},

		// lookup ke clients
		{{Key: "$lookup", Value: bson.M{
			"from":         "clients", // ✅ nama collection clients
			"localField":   "client_uuid",
			"foreignField": "uuid", // ✅ field uuid di clients
			"as":           "client",
		}}},

		// unwind supaya client jadi object
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$client",
			"preserveNullAndEmptyArrays": true, // kalau ada data orphan, tetap tampil
		}}},

		// sort
		{{Key: "$sort", Value: sort}},

		// pagination
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: int64(size)}},
	}

	log.Printf("Aggregation Pipeline: %+v\n", pipeline)

	cur, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []dao.ClientSubscriptionJoined
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *ClientSubscriptionRepositoryImpl) IsClientAllowed(clientUUID string) (bool, *dao.ClientSubscription, error) {
	sub, err := r.GetActiveByClientUUID(clientUUID)
	if err != nil {
		return false, nil, err
	}

	now := time.Now()
	if sub.IsActive && (sub.ExpiredAt.IsZero() || now.Before(sub.ExpiredAt)) {
		return true, sub, nil
	}

	return false, sub, nil
}

func ClientSubscriptionRepositoryInit(mongoClient *mongo.Client) *ClientSubscriptionRepositoryImpl {
	dbName := helpers.ProvideDBName()
	col := mongoClient.Database(dbName).Collection("client_subscriptions")
	subMasterCol := mongoClient.Database(dbName).Collection("subscriptions")
	return &ClientSubscriptionRepositoryImpl{col: col, subMasterCol: subMasterCol}
}

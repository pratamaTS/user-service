package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/helpers"
)

type ClientSubscriptionRepository interface {
	ActivateFromMaster(clientUUID, subscriptionUUID, createdBy string) (dao.ClientSubscription, error)
	GetActiveByClientUUID(clientUUID string) (*dao.ClientSubscription, error)
	ListClientBySubscriptionUUID(subscriptionUUID string) ([]dao.ClientSubscription, error)
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

func (r *ClientSubscriptionRepositoryImpl) ListClientBySubscriptionUUID(subscriptionUUID string) ([]dao.ClientSubscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if subscriptionUUID == "" {
		return nil, errors.New("subscription_uuid required")
	}

	cur, err := r.col.Find(ctx, bson.M{"subscription_uuid": subscriptionUUID}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []dao.ClientSubscription
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

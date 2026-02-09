package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"harjonan.id/user-service/app/helpers"
)

type RateLimitRepository interface {
	CheckRateLimit(ctx context.Context, key string, maxAttempts int, window time.Duration) error
}

type RateLimitRepositoryImpl struct {
	rateLimitCollection *mongo.Collection
}

func (u *RateLimitRepositoryImpl) CheckRateLimit(ctx context.Context, key string, maxAttempts int, window time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()

	var rl struct {
		Key       string    `bson:"key"`
		Count     int       `bson:"count"`
		ExpiresAt time.Time `bson:"expires_at"`
	}

	err := u.rateLimitCollection.FindOne(ctx, bson.M{"key": key}).Decode(&rl)
	if err == mongo.ErrNoDocuments {
		doc := bson.M{
			"key":        key,
			"count":      1,
			"expires_at": now.Add(window),
		}
		_, _ = u.rateLimitCollection.InsertOne(ctx, doc)
		return nil
	}

	if rl.ExpiresAt.Before(now) {
		_, _ = u.rateLimitCollection.UpdateOne(
			ctx,
			bson.M{"key": key},
			bson.M{
				"$set": bson.M{
					"count":      1,
					"expires_at": now.Add(window),
				},
			},
		)
		return nil
	}

	if rl.Count >= maxAttempts {
		return errors.New("too many attempts, please wait before trying again")
	}

	_, _ = u.rateLimitCollection.UpdateOne(
		ctx,
		bson.M{"key": key},
		bson.M{"$inc": bson.M{"count": 1}},
	)

	return nil

}

func RateLimitRepositoryInit(mongoClient *mongo.Client) *RateLimitRepositoryImpl {
	dbName := helpers.ProvideDBName()
	collection := mongoClient.Database(dbName).Collection("rate_limits")
	return &RateLimitRepositoryImpl{
		rateLimitCollection: collection,
	}
}

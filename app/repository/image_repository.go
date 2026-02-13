package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/helpers"
)

type ImageRepository interface {
	Save(img *dao.Image) (dao.Image, error)
	FindByKey(key string) (dao.Image, error)
}

type ImageRepositoryImpl struct {
	col *mongo.Collection
}

func ImageRepositoryInit(mongoClient *mongo.Client) *ImageRepositoryImpl {
	dbName := helpers.ProvideDBName()
	return &ImageRepositoryImpl{
		col: mongoClient.Database(dbName).Collection("images"),
	}
}

func (r *ImageRepositoryImpl) Save(img *dao.Image) (dao.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.col.InsertOne(ctx, img)
	if err != nil {
		return dao.Image{}, err
	}
	return *img, nil
}

func (r *ImageRepositoryImpl) FindByKey(key string) (dao.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var out dao.Image
	err := r.col.FindOne(ctx, bson.M{"key": key}).Decode(&out)
	return out, err
}

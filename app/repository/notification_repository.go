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

type NotificationRepository interface {
	Insert(n *dao.Notification) (dao.Notification, error)
	List(req *dto.FilterRequest) ([]dao.Notification, int64, error)

	MarkRead(uuid, clientUUID string) error
	MarkReadAll(clientUUID, branchUUID string) (int64, error)

	Clear(clientUUID, branchUUID string) (int64, error)
	ClearAll(clientUUID string) (int64, error)

	ExistsByRef(clientUUID string, ref string) (bool, error)
}

type NotificationRepositoryImpl struct {
	col *mongo.Collection
}

func NotificationRepositoryInit(mongoClient *mongo.Client) *NotificationRepositoryImpl {
	dbName := helpers.ProvideDBName()
	return &NotificationRepositoryImpl{
		col: mongoClient.Database(dbName).Collection("notifications"),
	}
}

func (r *NotificationRepositoryImpl) Insert(n *dao.Notification) (dao.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if n == nil {
		return dao.Notification{}, errors.New("payload required")
	}
	if n.ClientUUID == "" {
		return dao.Notification{}, errors.New("client_uuid required")
	}

	// ✅ penting: notif boleh branch-scope ATAU user-scope
	if n.BranchUUID == "" && n.UserUUID == "" {
		return dao.Notification{}, errors.New("branch_uuid or user_uuid required")
	}

	if n.Title == "" {
		return dao.Notification{}, errors.New("title required")
	}
	if n.Message == "" {
		return dao.Notification{}, errors.New("message required")
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	if n.UUID == "" {
		n.UUID = helpers.GenerateUUID()
	}
	n.IsRead = false
	n.CreatedAt = now
	n.CreatedAtStr = nowStr

	if _, err := r.col.InsertOne(ctx, n); err != nil {
		return dao.Notification{}, err
	}

	var out dao.Notification
	if err := r.col.FindOne(ctx, bson.M{"uuid": n.UUID}).Decode(&out); err != nil {
		return dao.Notification{}, err
	}
	return out, nil
}

func (r *NotificationRepositoryImpl) List(req *dto.FilterRequest) ([]dao.Notification, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{}

	// search
	if req != nil && req.Search != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": req.Search, "$options": "i"}},
			{"message": bson.M{"$regex": req.Search, "$options": "i"}},
			{"type": bson.M{"$regex": req.Search, "$options": "i"}},
			{"ref": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	// filter_by (support operator: $or, $in, dll)
	if req != nil {
		for k, v := range req.FilterBy {
			if k == "" || v == nil {
				continue
			}
			filter[k] = v
		}
	}

	// sort
	sort := bson.D{{Key: "created_at", Value: -1}}
	if req != nil {
		for k, v := range req.SortBy {
			if s, ok := v.(string); ok && (s == "asc" || s == "ASC" || s == "1") {
				sort = bson.D{{Key: k, Value: 1}}
			} else {
				sort = bson.D{{Key: k, Value: -1}}
			}
		}
	}

	// pagination
	page := 1
	size := 20
	if req != nil {
		page = req.Pagination.Page
		size = req.Pagination.PageSize
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	skip := int64((page - 1) * size)

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(sort).
		SetSkip(skip).
		SetLimit(int64(size))

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var out []dao.Notification
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

func (r *NotificationRepositoryImpl) MarkRead(uuid, clientUUID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if uuid == "" {
		return errors.New("uuid required")
	}
	if clientUUID == "" {
		return errors.New("client_uuid required")
	}

	_, err := r.col.UpdateOne(ctx,
		bson.M{"uuid": uuid, "client_uuid": clientUUID},
		bson.M{"$set": bson.M{"is_read": true}},
	)
	return err
}

// MarkReadAll:
// - kalau branchUUID diisi => mark read semua notif untuk branch itu (branch scope)
// - kalau branchUUID kosong => mark read semua notif 1 client (owner convenience)
func (r *NotificationRepositoryImpl) MarkReadAll(clientUUID, branchUUID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if clientUUID == "" {
		return 0, errors.New("client_uuid required")
	}

	filter := bson.M{
		"client_uuid": clientUUID,
		"is_read":     false,
	}
	if branchUUID != "" {
		filter["branch_uuid"] = branchUUID
	}

	res, err := r.col.UpdateMany(ctx, filter, bson.M{"$set": bson.M{"is_read": true}})
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

// Clear:
// - kalau branchUUID diisi => delete notif branch tsb
// - kalau branchUUID kosong => delete notif 1 client (owner convenience)
func (r *NotificationRepositoryImpl) Clear(clientUUID, branchUUID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if clientUUID == "" {
		return 0, errors.New("client_uuid required")
	}

	filter := bson.M{"client_uuid": clientUUID}
	if branchUUID != "" {
		filter["branch_uuid"] = branchUUID
	}

	res, err := r.col.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

func (r *NotificationRepositoryImpl) ClearAll(clientUUID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if clientUUID == "" {
		return 0, errors.New("client_uuid required")
	}

	res, err := r.col.DeleteMany(ctx, bson.M{"client_uuid": clientUUID})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

func (r *NotificationRepositoryImpl) ExistsByRef(clientUUID string, ref string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	if clientUUID == "" || ref == "" {
		return false, errors.New("client_uuid and ref required")
	}

	n, err := r.col.CountDocuments(ctx, bson.M{
		"client_uuid": clientUUID,
		"ref":         ref,
	})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

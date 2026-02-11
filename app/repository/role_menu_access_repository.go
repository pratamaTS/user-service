package repository

import (
	"context"
	"errors"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
)

type RoleMenuAccessRepository interface {
	SaveRoleMenuAccess(data *dao.RoleMenuAccess) (dao.RoleMenuAccess, error)
	GetMenusByRole(ctx context.Context, roleUUID string) ([]dto.UserMenu, error)
	DetailRoleMenuAccess(uuid string) (dao.RoleMenuAccess, error)
	ListRoleMenuAccess(req *dto.FilterRequest) ([]dao.RoleMenuAccess, error)
	DeleteRoleMenuAccess(uuid string) error
}

type RoleMenuAccessRepositoryImpl struct {
	roleMenuAccessCollection *mongo.Collection
	menuCollection           *mongo.Collection
}

func (u *RoleMenuAccessRepositoryImpl) SaveRoleMenuAccess(data *dao.RoleMenuAccess) (dao.RoleMenuAccess, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	switch {
	case data.UUID != "":
		filter["uuid"] = data.UUID
	case data.RoleUUID != "":
		filter["role_uuid"] = data.RoleUUID
	default:
		return dao.RoleMenuAccess{}, errors.New("upsert requires one of uuid/id/role_uuid")
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	data.UUID = helpers.GenerateUUID()
	data.UpdatedAt = now.Unix()
	data.UpdatedAtStr = nowStr

	update := bson.M{
		"$set": bson.M{
			"role_uuid":        data.RoleUUID,
			"accessible_menus": data.AccesibleMenu,
			"updated_at":       data.UpdatedAt,
			"updated_at_str":   data.UpdatedAtStr,
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

	if _, err := u.roleMenuAccessCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return dao.RoleMenuAccess{}, err
	}

	var out dao.RoleMenuAccess
	if err := u.roleMenuAccessCollection.FindOne(ctx, filter).Decode(&out); err != nil {
		return dao.RoleMenuAccess{}, err
	}
	return out, nil
}

func (u *RoleMenuAccessRepositoryImpl) GetMenusByRole(ctx context.Context, roleUUID string) ([]dto.UserMenu, error) {
	var rma dao.RoleMenuAccess
	if err := u.roleMenuAccessCollection.FindOne(ctx, bson.M{
		"role_uuid": roleUUID,
	}).Decode(&rma); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []dto.UserMenu{}, nil
		}
		return nil, err
	}

	if len(rma.AccesibleMenu) == 0 {
		return []dto.UserMenu{}, nil
	}

	menuUUIDs := make([]string, 0, len(rma.AccesibleMenu))
	accessMap := make(map[string][]string, len(rma.AccesibleMenu))

	for _, am := range rma.AccesibleMenu {
		menuUUIDs = append(menuUUIDs, am.MenuUUID)
		accessMap[am.MenuUUID] = am.AccessFunction
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"uuid": bson.M{"$in": menuUUIDs},
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "parent_menus",
			"localField":   "parent_uuid",
			"foreignField": "uuid",
			"as":           "parent",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$parent",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$sort", Value: bson.M{
			"sort": 1,
		}}},
	}

	cur, err := u.menuCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var menus []dto.UserMenu
	if err := cur.All(ctx, &menus); err != nil {
		return nil, err
	}

	for i := range menus {
		menus[i].AccessFunction = accessMap[menus[i].UUID]
	}

	sort.Slice(menus, func(i, j int) bool {
		return menus[i].SortOrder < menus[j].SortOrder
	})

	return menus, nil
}

func (u *RoleMenuAccessRepositoryImpl) DetailRoleMenuAccess(uuid string) (dao.RoleMenuAccess, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result dao.RoleMenuAccess
	err := u.roleMenuAccessCollection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&result)
	return result, err
}

func (u *RoleMenuAccessRepositoryImpl) ListRoleMenuAccess(req *dto.FilterRequest) ([]dao.RoleMenuAccess, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"role_uuid": bson.M{"$regex": req.Search, "$options": "i"}},
			{"menu.title": bson.M{"$regex": req.Search, "$options": "i"}},
			{"menu.href": bson.M{"$regex": req.Search, "$options": "i"}},
			{"menu.owner": bson.M{"$regex": req.Search, "$options": "i"}},
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
		sort = bson.D{{Key: "created_at", Value: -1}}
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

	cur, err := u.roleMenuAccessCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []dao.RoleMenuAccess
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (u *RoleMenuAccessRepositoryImpl) DeleteRoleMenuAccess(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := u.roleMenuAccessCollection.DeleteOne(ctx, bson.M{"uuid": uuid})
	return err
}

func RoleMenuAccessRepositoryInit(mongoClient *mongo.Client) *RoleMenuAccessRepositoryImpl {
	dbName := helpers.ProvideDBName()
	collection := mongoClient.Database(dbName).Collection("role_menu_access")
	menuCollection := mongoClient.Database(dbName).Collection("menus")
	return &RoleMenuAccessRepositoryImpl{
		roleMenuAccessCollection: collection,
		menuCollection:           menuCollection,
	}
}

package repository

import (
	"context"
	"errors"
	"sort"
	"strings"
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
	ListRoleMenuAccess(req *dto.FilterRequest) ([]dao.ResponseRoleMenuAccess, error)
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

func (u *RoleMenuAccessRepositoryImpl) ListRoleMenuAccess(req *dto.FilterRequest) ([]dao.ResponseRoleMenuAccess, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// -------------------------
	// FILTER (pre-lookup)
	// -------------------------
	filter := bson.M{}

	// search & filter untuk field root (role_uuid, dll) boleh sebelum unwind
	search := strings.TrimSpace(req.Search)
	if search != "" {
		filter["role_uuid"] = bson.M{"$regex": search, "$options": "i"}
	}

	for key, val := range req.FilterBy {
		if key == "" || val == nil {
			continue
		}
		// Filter untuk menu.* kita apply setelah lookup
		if strings.HasPrefix(key, "menu.") || strings.HasPrefix(key, "accessible_menus.menu.") {
			continue
		}
		filter[key] = val
	}

	// -------------------------
	// SORT
	// -------------------------
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

	// -------------------------
	// PAGINATION
	// -------------------------
	page := req.Pagination.Page
	size := req.Pagination.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	skip := int64((page - 1) * size)

	// -------------------------
	// PIPELINE
	// -------------------------
	pipeline := mongo.Pipeline{}

	// match root
	if len(filter) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: filter}})
	}

	// sort + paging di level dokumen root (sebelum unwind biar paging-nya bener per RoleMenuAccess)
	pipeline = append(pipeline,
		bson.D{{Key: "$sort", Value: sort}},
		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: int64(size)}},
	)

	// unwind accessible_menus (tiap item jadi 1 dokumen)
	pipeline = append(pipeline, bson.D{{Key: "$unwind", Value: bson.M{
		"path":                       "$accessible_menus",
		"preserveNullAndEmptyArrays": true,
	}}})

	// lookup menus berdasarkan accessible_menus.menu_uuid -> menus.uuid
	pipeline = append(pipeline,
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "menus",
			"localField":   "accessible_menus.menu_uuid",
			"foreignField": "uuid",
			"as":           "_menu",
		}}},
		bson.D{{Key: "$unwind", Value: bson.M{
			"path":                       "$_menu",
			"preserveNullAndEmptyArrays": true,
		}}},
	)

	// set hasil lookup ke accessible_menus.menu
	pipeline = append(pipeline, bson.D{{Key: "$set", Value: bson.M{
		"accessible_menus.menu": "$_menu",
	}}})

	// optional: match search untuk menu fields (setelah lookup)
	// (kalau mau search req.Search juga cari menu.title/href/owner)
	if search != "" {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.M{
			"$or": []bson.M{
				{"role_uuid": bson.M{"$regex": search, "$options": "i"}},
				{"accessible_menus.menu.title": bson.M{"$regex": search, "$options": "i"}},
				{"accessible_menus.menu.href": bson.M{"$regex": search, "$options": "i"}},
				{"accessible_menus.menu.owner": bson.M{"$regex": search, "$options": "i"}},
			},
		}}})
	}

	// apply filterBy untuk menu.* setelah lookup (kalau kamu memang pakai)
	postLookup := bson.M{}
	for key, val := range req.FilterBy {
		if key == "" || val == nil {
			continue
		}
		// dukung dua gaya key: "menu.title" atau "accessible_menus.menu.title"
		switch {
		case strings.HasPrefix(key, "menu."):
			postLookup["accessible_menus."+key] = val
		case strings.HasPrefix(key, "accessible_menus.menu."):
			postLookup[key] = val
		}
	}
	if len(postLookup) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: postLookup}})
	}

	// group balik jadi 1 dokumen per role_menu_access uuid
	pipeline = append(pipeline, bson.D{{Key: "$group", Value: bson.M{
		"_id":            "$uuid",
		"uuid":           bson.M{"$first": "$uuid"},
		"role_uuid":      bson.M{"$first": "$role_uuid"},
		"created_at":     bson.M{"$first": "$created_at"},
		"created_at_str": bson.M{"$first": "$created_at_str"},
		"updated_at":     bson.M{"$first": "$updated_at"},
		"updated_at_str": bson.M{"$first": "$updated_at_str"},
		"accessible_menus": bson.M{
			"$push": "$accessible_menus",
		},
	}}})

	// bersihin field helper _menu
	pipeline = append(pipeline, bson.D{{Key: "$project", Value: bson.M{
		"_id": 0,
	}}})

	cur, err := u.roleMenuAccessCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []dao.ResponseRoleMenuAccess
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

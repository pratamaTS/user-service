package repository

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
)

type AttendanceRepository interface {
	UpsertAttendance(req *dto.AttendanceUpsertRequest) (dao.Attendance, error)
	ListAttendance(req *dto.FilterRequest) ([]dao.Attendance, error)
}

type AttendanceRepositoryImpl struct {
	attendanceCol *mongo.Collection
	branchCol     *mongo.Collection
}

func AttendanceRepositoryInit(mongoClient *mongo.Client) *AttendanceRepositoryImpl {
	dbName := helpers.ProvideDBName()

	// ✅ sesuaikan koleksi branch kamu:
	branchCol := mongoClient.Database(dbName).Collection("client_branches")

	return &AttendanceRepositoryImpl{
		attendanceCol: mongoClient.Database(dbName).Collection("attendance_logs"),
		branchCol:     branchCol,
	}
}

func (r *AttendanceRepositoryImpl) UpsertAttendance(req *dto.AttendanceUpsertRequest) (dao.Attendance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	// basic validation
	if req.BranchUUID == "" {
		return dao.Attendance{}, errors.New("branch_uuid required")
	}
	if req.UserUUID == "" {
		return dao.Attendance{}, errors.New("user_uuid required")
	}
	if req.Type != string(dao.AttendanceCheckIn) && req.Type != string(dao.AttendanceCheckOut) {
		return dao.Attendance{}, errors.New("type must be CHECKIN or CHECKOUT")
	}
	if req.Status == "" {
		req.Status = string(dao.AttendancePresent)
	}

	// load branch
	var branch dao.ClientBranch
	if err := r.branchCol.FindOne(ctx, bson.M{
		"uuid":      req.BranchUUID,
		"is_active": true,
	}).Decode(&branch); err != nil {
		return dao.Attendance{}, errors.New("branch not found / inactive")
	}

	// parse branch lat lng
	branchLat := parseFloatSafe(branch.Latitude)
	branchLng := parseFloatSafe(branch.Longitude)
	if branchLat == 0 && branchLng == 0 {
		return dao.Attendance{}, errors.New("branch location not set")
	}

	// calculate distance
	dist := haversineMeters(branchLat, branchLng, req.LocationLatitude, req.LocationLongitude)

	// validate radius
	maxRadius := branch.MaxRadius
	if maxRadius <= 0 {
		maxRadius = 100 // default safety kalau kosong
	}
	inRadius := dist <= float64(maxRadius)
	if !inRadius {
		return dao.Attendance{}, errors.New("out of radius")
	}

	// enforce staff limit (based on OPEN checkins)
	// open = is_open true
	openCount, err := r.attendanceCol.CountDocuments(ctx, bson.M{
		"branch_uuid": req.BranchUUID,
		"is_open":     true,
	})
	if err != nil {
		return dao.Attendance{}, err
	}

	if req.Type == string(dao.AttendanceCheckIn) {
		// prevent double checkin (user already open)
		existOpen := r.attendanceCol.FindOne(ctx, bson.M{
			"branch_uuid": req.BranchUUID,
			"user_uuid":   req.UserUUID,
			"is_open":     true,
		})
		if existOpen.Err() == nil {
			return dao.Attendance{}, errors.New("already checked in (open session exists)")
		}

		// capacity check: openCount < total_staff
		if branch.TotalStaff > 0 && openCount >= int64(branch.TotalStaff) {
			return dao.Attendance{}, errors.New("attendance full: total staff limit reached")
		}

		// create new attendance row (open)
		now := time.Now()
		nowStr := now.Format(time.RFC3339)
		newUUID := helpers.GenerateUUID()

		doc := dao.Attendance{
			BaseModelV2: dao.BaseModelV2{
				UUID:         newUUID,
				CreatedAt:    now,
				CreatedAtStr: nowStr,
				UpdatedAt:    nil,
				UpdatedAtStr: nil,
			},
			ClientUUID: branch.ClientUUID,
			BranchUUID: req.BranchUUID,
			UserUUID:   req.UserUUID,
			UserName:   req.UserName,
			Type:       dao.AttendanceCheckIn,
			Status:     dao.AttendanceStatus(req.Status),
			Note:       req.Note,
			Location: dao.AttendanceLocation{
				Latitude:   req.LocationLatitude,
				Longitude:  req.LocationLongitude,
				Address:    req.LocationAddress,
				DistanceM:  dist,
				InRadius:   inRadius,
				MaxRadiusM: maxRadius,
			},
			IsOpen: true,
		}

		if _, err := r.attendanceCol.InsertOne(ctx, doc); err != nil {
			return dao.Attendance{}, err
		}
		return doc, nil
	}

	// CHECKOUT: must have open record
	var open dao.Attendance
	if err := r.attendanceCol.FindOne(ctx, bson.M{
		"branch_uuid": req.BranchUUID,
		"user_uuid":   req.UserUUID,
		"is_open":     true,
	}, options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})).Decode(&open); err != nil {
		return dao.Attendance{}, errors.New("no open attendance found for checkout")
	}

	// close it by updating same record
	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	update := bson.M{
		"$set": bson.M{
			"type":           string(dao.AttendanceCheckOut),
			"status":         req.Status,
			"note":           req.Note,
			"is_open":        false,
			"updated_at":     now,
			"updated_at_str": nowStr,
			"location": bson.M{
				"latitude":     req.LocationLatitude,
				"longitude":    req.LocationLongitude,
				"address":      req.LocationAddress,
				"distance_m":   dist,
				"in_radius":    inRadius,
				"max_radius_m": maxRadius,
			},
		},
	}

	if _, err := r.attendanceCol.UpdateOne(ctx, bson.M{"uuid": open.UUID}, update); err != nil {
		return dao.Attendance{}, err
	}

	// reload
	var out dao.Attendance
	if err := r.attendanceCol.FindOne(ctx, bson.M{"uuid": open.UUID}).Decode(&out); err != nil {
		return dao.Attendance{}, err
	}
	return out, nil
}

func (r *AttendanceRepositoryImpl) ListAttendance(req *dto.FilterRequest) ([]dao.Attendance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	filter := bson.M{}

	// search basic
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"user_name": bson.M{"$regex": req.Search, "$options": "i"}},
			{"note": bson.M{"$regex": req.Search, "$options": "i"}},
			{"branch_uuid": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}
	for k, v := range req.FilterBy {
		if k == "" || v == nil {
			continue
		}
		filter[k] = v
	}

	sort := bson.D{{Key: "created_at", Value: -1}}
	page := req.Pagination.Page
	size := req.Pagination.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	skip := int64((page - 1) * size)

	opts := options.Find().SetSort(sort).SetSkip(skip).SetLimit(int64(size))

	cur, err := r.attendanceCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var listOut []dao.Attendance
	if err := cur.All(ctx, &listOut); err != nil {
		return nil, err
	}
	return listOut, nil
}

func parseFloatSafe(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// Haversine distance in meters
func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000.0 // meters
	toRad := func(x float64) float64 { return x * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

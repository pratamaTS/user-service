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
	"harjonan.id/user-service/app/helpers"
)

type DashboardRepository interface {
	// Branch helper (untuk owner/company: ambil semua branch di company)
	GetCompanyBranchUUIDs(clientUUID string) ([]string, error)

	CountActiveProducts(clientUUID string) (int64, error)
	CountStockRequestMonth(branchUUIDs []string, monthStart, monthEnd time.Time) (int64, error)
	CountStockRequestProcess(branchUUIDs []string) (int64, error)
	CountJobWaitingAccept(branchUUIDs []string) (int64, error)

	CountTransactionToday(branchUUIDs []string, dayStart, dayEnd time.Time) (int64, error)
	SumRevenueMonth(branchUUIDs []string, monthStart, monthEnd time.Time) (float64, error)

	CountLowStockSKU(branchUUIDs []string, threshold int64) (int64, error)

	CountDriverAvailable(companyUUID string) (int64, error)
}

type DashboardRepositoryImpl struct {
	productCol       *mongo.Collection
	stockTransferCol *mongo.Collection
	posTrxCol        *mongo.Collection
	userCol          *mongo.Collection
	branchCol        *mongo.Collection
	roleCol          *mongo.Collection
}

func DashboardRepositoryInit(mongoClient *mongo.Client) *DashboardRepositoryImpl {
	dbName := helpers.ProvideDBName()

	productCol := mongoClient.Database(dbName).Collection("products")
	stockTransferCol := mongoClient.Database(dbName).Collection("stock_transfers")
	posTrxCol := mongoClient.Database(dbName).Collection("pos_transactions")
	userCol := mongoClient.Database(dbName).Collection("client_users")
	branchCol := mongoClient.Database(dbName).Collection("client_branches")
	roleCol := mongoClient.Database(dbName).Collection("roles")

	return &DashboardRepositoryImpl{
		productCol:       productCol,
		stockTransferCol: stockTransferCol,
		posTrxCol:        posTrxCol,
		userCol:          userCol,
		branchCol:        branchCol,
		roleCol:          roleCol,
	}
}

// ----------------------------------------------------
// Branch UUID list for company (owner/company scope)
// ----------------------------------------------------
func (r *DashboardRepositoryImpl) GetCompanyBranchUUIDs(clientUUID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if clientUUID == "" {
		return nil, errors.New("client_uuid required")
	}

	// ✅ asumsi dokumen branch punya field company_uuid & uuid
	filter := bson.M{
		"client_uuid": clientUUID,
		"is_active":   true,
	}

	opts := options.Find().SetProjection(bson.M{"uuid": 1})
	cur, err := r.branchCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var rows []bson.M
	if err := cur.All(ctx, &rows); err != nil {
		return nil, err
	}

	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if v, ok := row["uuid"].(string); ok && v != "" {
			out = append(out, v)
		}
	}
	return out, nil
}

// ----------------------------------------------------
// Products
// ----------------------------------------------------
func (r *DashboardRepositoryImpl) CountActiveProducts(clientUUID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{
		"is_active": true,
	}
	var clientBranch dao.ClientBranch
	err := r.branchCol.FindOne(ctx, bson.M{"client_uuid": clientUUID, "name": bson.M{"$regex": "Gudang", "$options": "i"}}).Decode(&clientBranch)
	if err != nil {
		log.Printf("CountActiveProducts: failed to find client branch for client_uuid=%s: %v", clientUUID, err)
		return 0, err
	}

	filter["branch_uuid"] = clientBranch.UUID

	totalProducts, err := r.productCol.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("CountActiveProducts: failed to count products for client_uuid=%s: %v", clientUUID, err)
		return 0, err
	}

	log.Printf("CountActiveProducts: client_uuid=%s, branch_uuid=%s, total_active_products=%d", clientUUID, clientBranch.UUID, totalProducts)

	return totalProducts, nil
}

func (r *DashboardRepositoryImpl) CountLowStockSKU(branchUUIDs []string, threshold int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{
		"is_active": true,
		"stock":     bson.M{"$lte": threshold},
	}
	if len(branchUUIDs) > 0 {
		filter["branch_uuid"] = bson.M{"$in": branchUUIDs}
	}

	totalLowStock, err := r.productCol.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("CountLowStockSKU: failed to count low stock products for branchUUIDs=%v: %v", branchUUIDs, err)
		return 0, err
	}

	return totalLowStock, nil
}

// ----------------------------------------------------
// Stock Transfer / Stock Request
// ----------------------------------------------------
func (r *DashboardRepositoryImpl) CountStockRequestMonth(branchUUIDs []string, monthStart, monthEnd time.Time) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{
		"created_at": bson.M{"$gte": monthStart, "$lt": monthEnd},
	}
	if len(branchUUIDs) > 0 {
		filter["$or"] = []bson.M{
			{"from_branch_uuid": bson.M{"$in": branchUUIDs}},
			{"to_branch_uuid": bson.M{"$in": branchUUIDs}},
		}
	}

	totalStockReq, err := r.stockTransferCol.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("CountStockRequestMonth: failed to count stock requests for branchUUIDs=%v, monthStart=%s, monthEnd=%s: %v", branchUUIDs, monthStart.Format(time.RFC3339), monthEnd.Format(time.RFC3339), err)
		return 0, err
	}

	return totalStockReq, nil
}

func (r *DashboardRepositoryImpl) CountStockRequestProcess(branchUUIDs []string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{
		"status": bson.M{"$in": []string{
			"PENDING_WAREHOUSE",
			"WAITING_DRIVER",
			"IN_PROGRESS",
		}},
	}
	if len(branchUUIDs) > 0 {
		filter["$or"] = []bson.M{
			{"from_branch_uuid": bson.M{"$in": branchUUIDs}},
			{"to_branch_uuid": bson.M{"$in": branchUUIDs}},
		}
	}

	totalProcess, err := r.stockTransferCol.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("CountStockRequestProcess: failed to count stock requests for branchUUIDs=%v: %v", branchUUIDs, err)
		return 0, err
	}

	return totalProcess, nil
}

func (r *DashboardRepositoryImpl) CountJobWaitingAccept(branchUUIDs []string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{
		"status": "WAITING_DRIVER",
	}
	if len(branchUUIDs) > 0 {
		filter["$or"] = []bson.M{
			{"from_branch_uuid": bson.M{"$in": branchUUIDs}},
			{"to_branch_uuid": bson.M{"$in": branchUUIDs}},
		}
	}

	totalWaitingAccept, err := r.stockTransferCol.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("CountJobWaitingAccept: failed to count waiting accept jobs for branchUUIDs=%v: %v", branchUUIDs, err)
		return 0, err
	}

	return totalWaitingAccept, nil
}

// ----------------------------------------------------
// POS Transactions (Revenue)
// ----------------------------------------------------
func (r *DashboardRepositoryImpl) CountTransactionToday(branchUUIDs []string, dayStart, dayEnd time.Time) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{
		"status":     "PAID",
		"created_at": bson.M{"$gte": dayStart, "$lt": dayEnd},
	}
	if len(branchUUIDs) > 0 {
		filter["branch_uuid"] = bson.M{"$in": branchUUIDs}
	}

	totalTransactions, err := r.posTrxCol.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("CountTransactionToday: failed to count transactions for branchUUIDs=%v, dayStart=%s, dayEnd=%s: %v", branchUUIDs, dayStart.Format(time.RFC3339), dayEnd.Format(time.RFC3339), err)
		return 0, err
	}

	return totalTransactions, nil
}

func (r *DashboardRepositoryImpl) SumRevenueMonth(branchUUIDs []string, monthStart, monthEnd time.Time) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	match := bson.M{
		"status":     "PAID",
		"created_at": bson.M{"$gte": monthStart, "$lt": monthEnd},
	}
	if len(branchUUIDs) > 0 {
		match["branch_uuid"] = bson.M{"$in": branchUUIDs}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$group", Value: bson.M{
			"_id": nil,
			"sum": bson.M{"$sum": "$total"}, // ✅ sesuai dao.POSTransaction.Total
		}}},
	}

	cur, err := r.posTrxCol.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("SumRevenueMonth: failed to aggregate revenue for branchUUIDs=%v, monthStart=%s, monthEnd=%s: %v", branchUUIDs, monthStart.Format(time.RFC3339), monthEnd.Format(time.RFC3339), err)
		return 0, err
	}
	defer cur.Close(ctx)

	var out []bson.M
	if err := cur.All(ctx, &out); err != nil {
		return 0, err
	}
	if len(out) == 0 {
		return 0, nil
	}

	switch v := out[0]["sum"].(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	default:
		return 0, nil
	}
}

// ----------------------------------------------------
// Drivers
// ----------------------------------------------------
func (r *DashboardRepositoryImpl) CountDriverAvailable(clientUUID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if clientUUID == "" {
		return 0, errors.New("client_uuid required")
	}

	// 1) ambil role DRIVER
	var roleDoc dao.Role
	if err := r.roleCol.FindOne(ctx, bson.M{"value": "DRIVER"}).Decode(&roleDoc); err != nil {
		log.Printf("CountDriverAvailable: failed to find DRIVER role: %v", err)
		return 0, err
	}

	// 2) ambil semua user DRIVER di client tsb
	cur, err := r.userCol.Find(ctx, bson.M{
		"client_uuid": clientUUID,
		"role_uuid":   roleDoc.UUID,
	})
	if err != nil {
		log.Printf("CountDriverAvailable: failed to find drivers for client_uuid=%s role_uuid=%s: %v", clientUUID, roleDoc.UUID, err)
		return 0, err
	}
	defer cur.Close(ctx)

	var drivers []dao.ClientUser
	if err := cur.All(ctx, &drivers); err != nil {
		log.Printf("CountDriverAvailable: failed to decode drivers for client_uuid=%s role_uuid=%s: %v", clientUUID, roleDoc.UUID, err)
		return 0, err
	}

	if len(drivers) == 0 {
		return 0, nil
	}

	// 3) driver dianggap "available" jika TIDAK punya job dengan status IN_PROGRESS / WAITING_DRIVER
	//    (sesuaikan list status aktif kamu kalau beda)
	busyStatuses := []string{
		string(dao.StockTransferInProgress),
	}

	totalAvailable := int64(0)

	for _, driver := range drivers {
		driverUUID := driver.UserUUID
		if driverUUID == "" {
			continue
		}

		busyCount, err := r.stockTransferCol.CountDocuments(ctx, bson.M{
			"driver_uuid": driverUUID,
			"status":      bson.M{"$in": busyStatuses},
		})
		if err != nil {
			log.Printf("CountDriverAvailable: failed counting busy jobs for driver_uuid=%s: %v", driverUUID, err)
			return 0, err
		}

		if busyCount == 0 {
			totalAvailable++
		}
	}

	return totalAvailable, nil
}

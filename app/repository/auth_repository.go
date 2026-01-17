package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"harjonan.id/user-service/app/constant"
	"harjonan.id/user-service/app/domain/dao"
	"harjonan.id/user-service/app/domain/dto"
	"harjonan.id/user-service/app/helpers"
)

type AuthRepository interface {
	Register(data map[string]any, subject string) (dao.Auth, error)
	Login(username, password string) (dao.Auth, error)
	ValidateToken(accessToken string) (*dto.UserProfile, error)
	Refresh(accessToken, refreshToken string) (dao.Auth, error)
	Logout(accessToken string) error
}

type AuthRepositoryImpl struct {
	authCollection        *mongo.Collection
	adminCollection       *mongo.Collection
	userCollection        *mongo.Collection
	clientCollection      *mongo.Collection
	companyCollection     *mongo.Collection
	roleCollection        *mongo.Collection
	clientUserCollection  *mongo.Collection
	companyUserCollection *mongo.Collection
}

func AuthRepositoryInit(mc *mongo.Client) *AuthRepositoryImpl {
	db := mc.Database("db_portal_general")
	return &AuthRepositoryImpl{
		authCollection:        db.Collection("cl_auths"),
		adminCollection:       db.Collection("cl_admins"),
		userCollection:        db.Collection("cl_users"),
		clientCollection:      db.Collection("cl_clients"),
		companyCollection:     db.Collection("cl_companies"),
		roleCollection:        db.Collection("cl_roles"),
		clientUserCollection:  db.Collection("cl_client_users"),
		companyUserCollection: db.Collection("cl_company_users"),
	}
}

func (r *AuthRepositoryImpl) Register(data map[string]any, subject string) (dao.Auth, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name, _ := data["name"].(string)
	email, _ := data["email"].(string)
	username, _ := data["username"].(string)
	password, _ := data["password"].(string)
	phone, _ := data["phone_number"].(string)
	address, _ := data["address"].(string)
	image, _ := data["image"].(string)

	if email == "" || password == "" {
		return dao.Auth{}, errors.New("email and password required")
	}
	if subject == "" {
		subject = "CLIENT"
	}

	// hash password
	hasher := sha256.New()
	hasher.Write([]byte(password))
	hashedPass := hex.EncodeToString(hasher.Sum(nil))

	uuid := helpers.GenerateUUID()

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	switch subject {
	case "COMPANY":
		admin := dao.Admin{
			BaseModel: dao.BaseModel{
				UUID:         uuid,
				CreatedAt:    now,
				CreatedAtStr: nowStr,
				UpdatedAt:    now.Unix(),
				UpdatedAtStr: nowStr,
			},
			Image:       image,
			Username:    username,
			Name:        name,
			Email:       email,
			Password:    hashedPass,
			PhoneNumber: phone,
			Address:     address,
			IsActive:    true,
		}
		if _, err := r.adminCollection.InsertOne(ctx, admin); err != nil {
			return dao.Auth{}, err
		}
	case "CLIENT":
		user := dao.User{
			BaseModel: dao.BaseModel{
				UUID:         uuid,
				CreatedAt:    now,
				CreatedAtStr: nowStr,
				UpdatedAt:    now.Unix(),
				UpdatedAtStr: nowStr,
			},
			Image:       image,
			Username:    username,
			Name:        name,
			Email:       email,
			Password:    hashedPass,
			PhoneNumber: phone,
			Address:     address,
			IsActive:    true,
		}
		if _, err := r.userCollection.InsertOne(ctx, user); err != nil {
			return dao.Auth{}, err
		}
	default:
		return dao.Auth{}, errors.New("unknown subject type")
	}

	auth, err := r.Login(username, password)
	if err != nil {
		return dao.Auth{}, err
	}

	return auth, nil
}

// =====================================================
// LOGIN
// =====================================================

func (r *AuthRepositoryImpl) Login(username, password string) (dao.Auth, error) {
	if username == "" || password == "" {
		return dao.Auth{}, errors.New("username and password required")
	}

	hasher := sha256.New()
	hasher.Write([]byte(password))
	hashedPass := hex.EncodeToString(hasher.Sum(nil))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var userUUID string
	var admin dao.Admin
	var user dao.User
	var subject string

	if err := r.adminCollection.FindOne(ctx, bson.M{"username": username, "password": hashedPass, "is_active": true}).Decode(&admin); err == nil {
		userUUID = admin.UUID
		subject = "COMPANY"
	} else {
		if err := r.userCollection.FindOne(ctx, bson.M{"username": username, "password": hashedPass, "is_active": true}).Decode(&user); err != nil {
			return dao.Auth{}, errors.New("invalid username or password")
		}
		userUUID = user.UUID
		subject = "CLIENT"
	}

	sessionID := helpers.GenerateUUID()
	jwtSecret := helpers.ProvideJWTSecret()
	accessTTL := helpers.ProvideAccessTTL()
	refreshTTL := helpers.ProvideRefreshTTL()

	access, accessExp, err := helpers.GenerateJWT(
		subject,
		jwtSecret,
		accessTTL,
		helpers.TokenTypeAccess,
		sessionID,
	)
	if err != nil {
		return dao.Auth{}, err
	}

	refresh, refreshExp, err := helpers.GenerateJWT(
		subject,
		jwtSecret,
		refreshTTL,
		helpers.TokenTypeRefresh,
		sessionID,
	)
	if err != nil {
		return dao.Auth{}, err
	}

	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	doc := bson.M{
		"user_uuid":          userUUID,
		"subject":            subject,
		"session_id":         sessionID,
		"token":              access,
		"expired_at":         accessExp,
		"refresh_token":      refresh,
		"refresh_expired_at": refreshExp,
		"login_at":           now.Unix(),
		"status":             constant.LOGIN,
		"updated_at":         now.Unix(),
		"updated_at_str":     nowStr,
	}

	_, err = r.authCollection.UpdateOne(
		context.Background(),
		bson.M{"token": access},
		bson.M{
			"$set":         doc,
			"$setOnInsert": bson.M{"created_at": now, "created_at_str": nowStr},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return dao.Auth{}, err
	}

	var out dao.Auth
	if err := r.authCollection.FindOne(context.Background(), bson.M{"token": access}).Decode(&out); err != nil {
		return dao.Auth{}, err
	}
	return out, nil
}

// =====================================================
// VALIDATE TOKEN
// =====================================================

func (r *AuthRepositoryImpl) ValidateToken(accessToken string) (*dto.UserProfile, error) {
	if accessToken == "" {
		return nil, errors.New("missing token")
	}

	jwtSecret := helpers.ProvideJWTSecret()

	tok, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	if tokenType, _ := claims["type"].(string); tokenType != "" && tokenType != string(helpers.TokenTypeAccess) {
		return nil, errors.New("invalid token type")
	}

	ctx := context.Background()

	var auth dao.Auth
	if err := r.authCollection.FindOne(ctx, bson.M{"token": accessToken}).Decode(&auth); err != nil {
		return nil, errors.New("token not found")
	}

	now := time.Now().Unix()
	if auth.Status == constant.LOGOUT {
		return nil, errors.New("token revoked")
	}
	if now > auth.ExpiredAt {
		_, _ = r.authCollection.UpdateOne(ctx,
			bson.M{"token": accessToken},
			bson.M{"$set": bson.M{
				"status":         constant.EXPIRED,
				"updated_at":     now,
				"updated_at_str": time.Now().Format(time.RFC3339),
			}},
		)
		return nil, errors.New("token expired")
	}

	userUUID := auth.UserUUID
	subject := auth.Subject

	log.Print("User UUID: ", userUUID)
	log.Print("Subject profile: ", subject)

	switch subject {
	case "COMPANY":
		return r.buildAdminProfile(userUUID)
	case "CLIENT":
		return r.buildClientUserProfile(userUUID)
	default:
		if p, err := r.buildClientUserProfile(userUUID); err == nil {
			return p, nil
		}
		return r.buildAdminProfile(userUUID)
	}
}

func (r *AuthRepositoryImpl) Refresh(accessToken, refreshToken string) (dao.Auth, error) {
	if accessToken == "" || refreshToken == "" {
		return dao.Auth{}, errors.New("missing tokens")
	}

	ctx := context.Background()

	var auth dao.Auth
	if err := r.authCollection.FindOne(ctx, bson.M{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"status":        constant.LOGIN,
	}).Decode(&auth); err != nil {
		return dao.Auth{}, errors.New("session not found or not active")
	}

	now := time.Now().Unix()
	if now > auth.RefreshExpiredAt {
		_, _ = r.authCollection.UpdateOne(ctx,
			bson.M{"token": accessToken},
			bson.M{"$set": bson.M{
				"status":         constant.EXPIRED,
				"updated_at":     now,
				"updated_at_str": time.Now().Format(time.RFC3339),
			}},
		)
		return dao.Auth{}, errors.New("refresh token expired")
	}

	jwtSecret := helpers.ProvideJWTSecret()

	rt, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !rt.Valid {
		return dao.Auth{}, errors.New("invalid refresh token")
	}
	if claims, ok := rt.Claims.(jwt.MapClaims); ok {
		if tokenType, _ := claims["type"].(string); tokenType != "" && tokenType != string(helpers.TokenTypeRefresh) {
			return dao.Auth{}, errors.New("invalid refresh token type")
		}
	}

	sessionID := auth.SessionID
	if sessionID == "" {
		sessionID = helpers.GenerateUUID()
	}

	accessTTL := helpers.ProvideAccessTTL()
	refreshTTL := helpers.ProvideRefreshTTL()

	newAccess, newAccessExp, err := helpers.GenerateJWT(
		auth.Subject,
		jwtSecret,
		accessTTL,
		helpers.TokenTypeAccess,
		sessionID,
	)
	if err != nil {
		return dao.Auth{}, err
	}

	newRefresh, newRefreshExp, err := helpers.GenerateJWT(
		auth.Subject,
		jwtSecret,
		refreshTTL,
		helpers.TokenTypeRefresh,
		sessionID,
	)
	if err != nil {
		return dao.Auth{}, err
	}

	nowTime := time.Now()
	nowStr := nowTime.Format(time.RFC3339)

	_, _ = r.authCollection.UpdateOne(ctx,
		bson.M{"token": accessToken},
		bson.M{"$set": bson.M{
			"status":         constant.EXPIRED,
			"updated_at":     now,
			"updated_at_str": nowStr,
		}},
	)

	newDoc := dao.Auth{
		BaseModel: dao.BaseModel{
			CreatedAt:    nowTime,
			CreatedAtStr: nowStr,
			UpdatedAt:    now,
			UpdatedAtStr: nowStr,
		},
		UserUUID:         auth.UserUUID,
		Subject:          auth.Subject,
		SessionID:        sessionID,
		Token:            newAccess,
		ExpiredAt:        newAccessExp,
		RefreshToken:     newRefresh,
		RefreshExpiredAt: newRefreshExp,
		LoginAt:          auth.LoginAt,
		Status:           constant.LOGIN,
	}

	_, err = r.authCollection.UpdateOne(
		ctx,
		bson.M{"token": newAccess},
		bson.M{
			"$set":         newDoc,
			"$setOnInsert": bson.M{"created_at": nowTime, "created_at_str": nowStr},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return dao.Auth{}, err
	}

	var out dao.Auth
	if err := r.authCollection.FindOne(ctx, bson.M{"token": newAccess}).Decode(&out); err != nil {
		return dao.Auth{}, err
	}
	return out, nil
}

// =====================================================
// LOGOUT
// =====================================================

func (r *AuthRepositoryImpl) Logout(accessToken string) error {
	if accessToken == "" {
		return errors.New("missing token")
	}
	_, err := r.authCollection.UpdateOne(
		context.Background(),
		bson.M{"token": accessToken},
		bson.M{"$set": bson.M{
			"status":         constant.LOGOUT,
			"updated_at":     time.Now().Unix(),
			"updated_at_str": time.Now().Format(time.RFC3339),
		}},
	)
	return err
}

// =====================================================
// BUILD PROFILE HELPERS
// =====================================================

func (r *AuthRepositoryImpl) buildAdminProfile(userUUID string) (*dto.UserProfile, error) {
	ctx := context.Background()

	var admin dao.Admin
	if err := r.adminCollection.FindOne(ctx, bson.M{"uuid": userUUID}).Decode(&admin); err != nil {
		return nil, errors.New("admin not found")
	}

	var company dao.Client
	var role dao.Role
	var mapDoc struct {
		CompanyUUID string `bson:"company_uuid"`
		RoleUUID    string `bson:"role_uuid"`
	}
	_ = r.companyUserCollection.FindOne(ctx, bson.M{"user_uuid": userUUID}).Decode(&mapDoc)
	if mapDoc.CompanyUUID != "" {
		_ = r.companyCollection.FindOne(ctx, bson.M{"uuid": mapDoc.CompanyUUID}).Decode(&company)
	}
	if mapDoc.RoleUUID != "" {
		_ = r.roleCollection.FindOne(ctx, bson.M{"uuid": mapDoc.RoleUUID}).Decode(&role)
	}

	return &dto.UserProfile{
		UUID:        admin.UUID,
		Client:      company,
		Role:        role,
		Username:    admin.Username,
		Name:        admin.Name,
		Email:       admin.Email,
		PhoneNumber: admin.PhoneNumber,
		Address:     admin.Address,
	}, nil
}

func (r *AuthRepositoryImpl) buildClientUserProfile(userUUID string) (*dto.UserProfile, error) {
	log.Print("User UUID: ", userUUID)
	ctx := context.Background()

	var user dao.User
	if err := r.userCollection.FindOne(ctx, bson.M{"uuid": userUUID}).Decode(&user); err != nil {
		return nil, errors.New("user not found")
	}

	var client dao.Client
	var role dao.Role
	var mapDoc struct {
		ClientUUID string `bson:"company_uuid"`
		RoleUUID   string `bson:"role_uuid"`
	}
	_ = r.clientUserCollection.FindOne(ctx, bson.M{"user_uuid": userUUID}).Decode(&mapDoc)
	log.Println("Mapping doc:", mapDoc)
	if mapDoc.ClientUUID != "" {
		log.Println("Fetching client for UUID:", mapDoc.ClientUUID)
		_ = r.clientCollection.FindOne(ctx, bson.M{"uuid": mapDoc.ClientUUID}).Decode(&client)
	}
	if mapDoc.RoleUUID != "" {
		_ = r.roleCollection.FindOne(ctx, bson.M{"uuid": mapDoc.RoleUUID}).Decode(&role)
	}

	return &dto.UserProfile{
		UUID:        user.UUID,
		Client:      client,
		Role:        role,
		Username:    user.Username,
		Name:        user.Name,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Address:     user.Address,
	}, nil
}

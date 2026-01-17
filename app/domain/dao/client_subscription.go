package dao

import "time"

type ClientSubscription struct {
	BaseModel `bson:",inline"`

	ClientUUID       string `json:"client_uuid" bson:"client_uuid" validate:"required"`
	SubscriptionUUID string `json:"subscription_uuid" bson:"subscription_uuid" validate:"required"` // master plan uuid

	StartAt   time.Time `json:"start_at" bson:"start_at"`
	ExpiredAt time.Time `json:"expired_at" bson:"expired_at"`

	// snapshot penting: biar kalau master berubah, record aktif tetap konsisten
	Type          SubscriptionType `json:"type" bson:"type"`
	BillingPeriod BillingPeriod    `json:"billing_period" bson:"billing_period"`
	IsFeatureFull bool             `json:"is_feature_full" bson:"is_feature_full"`
	MaxUser       int              `json:"max_user" bson:"max_user"`
	MaxProject    int              `json:"max_project" bson:"max_project"`
	MaxStorageGB  int              `json:"max_storage_gb" bson:"max_storage_gb"`

	IsActive  bool   `json:"is_active" bson:"is_active"`
	CreatedBy string `json:"created_by" bson:"created_by"`
}

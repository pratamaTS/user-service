package dao

type Subscription struct {
	BaseModel     `bson:",inline"`
	Name          string           `json:"name" bson:"name" validate:"required"`
	Type          SubscriptionType `json:"type" bson:"type" validate:"required"`
	BillingPeriod BillingPeriod    `json:"billing_period" bson:"billing_period"`
	DurationDays  int              `json:"duration_days" bson:"duration_days"`
	Price         float64          `json:"price" bson:"price"`
	IsFeatureFull bool             `json:"is_feature_full" bson:"is_feature_full"`
	MaxUser       int              `json:"max_user" bson:"max_user"`
	MaxProject    int              `json:"max_project" bson:"max_project"`
	MaxStorageGB  int              `json:"max_storage_gb" bson:"max_storage_gb"`
	IsActive      bool             `json:"is_active" bson:"is_active"`
	CreatedBy     string           `json:"created_by" bson:"created_by"`
}

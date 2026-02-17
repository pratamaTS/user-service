package dao

import "time"

type BaseModel struct {
	UUID         string    `bson:"uuid,omitempty" json:"uuid"`
	CreatedAt    time.Time `bson:"created_at,omitempty" json:"created_at"`
	CreatedAtStr string    `bson:"created_at_str,omitempty" json:"created_at_str"`
	UpdatedAt    int64     `bson:"updated_at,omitempty" json:"-"`
	UpdatedAtStr string    `bson:"updated_at_str,omitempty" json:"updated_at_str"`
}

type BaseModelV2 struct {
	UUID         string     `bson:"uuid,omitempty" json:"uuid"`
	CreatedAt    time.Time  `bson:"created_at,omitempty" json:"created_at"`
	CreatedAtStr string     `bson:"created_at_str,omitempty" json:"created_at_str"`
	UpdatedAt    *time.Time `bson:"updated_at,omitempty" json:"-"`
	UpdatedAtStr *string    `bson:"updated_at_str,omitempty" json:"updated_at_str"`
}

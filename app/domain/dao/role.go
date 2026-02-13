package dao

type Role struct {
	BaseModel `bson:",inline"`
	Name      string `bson:"name" json:"name" validate:"required"`
	Value     string `bson:"value" json:"value" validate:"required"`
	IsUseBy   string `json:"is_use_by" bson:"is_use_by" validate:"is_use_by,required"`
	IsActive  bool   `json:"is_active" bson:"is_active"`
}

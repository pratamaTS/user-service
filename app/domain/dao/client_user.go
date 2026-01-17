package dao

type ClientUser struct {
	BaseModel  `bson:",inline"`
	ClientUUID string `bson:"client_uuid" json:"client_uuid" validate:"required"`
	RoleUUID   string `json:"role_uuid" bson:"role_uuid" validate:"required"`
	UserUUID   string `json:"user_uuid" bson:"user_uuid" validate:"required"`
}

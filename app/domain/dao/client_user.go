package dao

type ClientUser struct {
	BaseModel  `bson:",inline"`
	ClientUUID string `bson:"client_uuid" json:"client_uuid" validate:"required"`
	BranchUUID string `bson:"branch_uuid" json:"branch_uuid"`
	RoleUUID   string `json:"role_uuid" bson:"role_uuid" validate:"required"`
	UserUUID   string `json:"user_uuid" bson:"user_uuid" validate:"required"`
}

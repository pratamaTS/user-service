package dao

type ClientUser struct {
	BaseModel  `bson:",inline"`
	ClientUUID string `bson:"client_uuid" json:"client_uuid" validate:"required"`
	BranchUUID string `bson:"branch_uuid" json:"branch_uuid"`
	RoleUUID   string `json:"role_uuid" bson:"role_uuid" validate:"required"`
	UserUUID   string `json:"user_uuid" bson:"user_uuid" validate:"required"`
}

type ClientUserWithDetail struct {
	BaseModel `bson:",inline"`

	ClientUUID string `bson:"client_uuid" json:"client_uuid"`
	BranchUUID string `bson:"branch_uuid" json:"branch_uuid"`
	RoleUUID   string `bson:"role_uuid" json:"role_uuid"`
	UserUUID   string `bson:"user_uuid" json:"user_uuid"`

	Client Client       `bson:"client,omitempty" json:"client,omitempty"`
	Branch ClientBranch `bson:"branch,omitempty" json:"branch,omitempty"`
	Role   Role         `bson:"role,omitempty" json:"role,omitempty"`
	User   User         `bson:"user,omitempty" json:"user,omitempty"`

	IsOwner bool `bson:"is_owner" json:"is_owner"`
}

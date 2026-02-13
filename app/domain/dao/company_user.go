package dao

type CompanyUser struct {
	BaseModel   `bson:",inline"`
	CompanyUUID string `bson:"company_uuid" json:"company_uuid" validate:"required"`
	RoleUUID    string `json:"role_uuid" bson:"role_uuid" validate:"required"`
	AdminUUID   string `json:"AdminUUID" bson:"UserUUID" validate:"required"`
}

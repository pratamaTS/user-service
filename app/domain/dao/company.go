package dao

type Company struct {
	BaseModel    `bson:",inline"`
	Logo         string `bson:"logo" json:"logo"`
	Name         string `json:"name" bson:"name" validate:"required"`
	Host         string `bson:"host" json:"host"`
	WebisteUrl   string `bson:"website_url" json:"website_url"`
	PhoneNumber  string `bson:"phone_number" json:"phone_number"`
	Address      string `json:"address" bson:"address"`
	Website      string `bson:"website" json:"website"`
	NIB          string `json:"nib" bson:"nib" validate:"required"`
	NPWP         string `json:"npwp" bson:"npwp" validate:"required"`
	EmailCompany string `json:"email_company" bson:"email_company" validate:"required"`
	EmailNotif   string `json:"email_notif" bson:"email_notif" validate:"required"`
	IsActive     bool   `json:"is_active" bson:"is_active"`
}

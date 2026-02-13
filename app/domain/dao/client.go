package dao

type Client struct {
	BaseModel   `bson:",inline"`
	CompanyUUID string `bson:"company_uuid" json:"company_uuid" validate:"required"`
	Logo        string `bson:"logo" json:"logo"`
	Name        string `json:"name" bson:"name" validate:"required"`
	Host        string `bson:"host" json:"host"`
	WebisteUrl  string `bson:"website_url" json:"website_url"`
	PhoneNumber string `bson:"phone_number" json:"phone_number"`
	IsActive    bool   `json:"is_active" bson:"is_active"`
	CreatedBy   string `json:"created_by" bson:"created_by"`
}

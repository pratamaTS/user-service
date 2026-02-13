package dao

type ClientBranch struct {
	BaseModel   `bson:",inline"`
	ClientUUID  string `bson:"client_uuid" json:"client_uuid" validate:"required"`
	Name        string `json:"name" bson:"name" validate:"required"`
	Address     string `bson:"address" json:"address"`
	Longitude   string `bson:"longitude" json:"longitude"`
	Latitude    string `bson:"latitude" json:"latitude"`
	PhoneNumber string `bson:"phone_number" json:"phone_number"`
	TotalStaff  int16  `bson:"total_staff" json:"total_staff"`
	MaxRadius   int16  `bson:"max_radius" json:"max_radius"`
	IsActive    bool   `json:"is_active" bson:"is_active"`
	CreatedBy   string `json:"created_by" bson:"created_by"`
}

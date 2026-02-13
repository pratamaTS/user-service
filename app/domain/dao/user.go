package dao

type User struct {
	BaseModel   `bson:",inline"`
	Image       string `bson:"image" json:"image"`
	Username    string `bson:"username" json:"username"`
	Name        string `bson:"name" json:"name"`
	Email       string `bson:"email" json:"email"`
	Password    string `bson:"password" json:"password"`
	PhoneNumber string `bson:"phone_number" json:"phone_number"`
	Address     string `bson:"address" json:"address"`
	IsActive    bool   `json:"is_active" bson:"is_active"`
}

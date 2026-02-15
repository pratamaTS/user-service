package dto

import "harjonan.id/user-service/app/domain/dao"

type UserProfile struct {
	UUID        string           `json:"uuid"`
	Client      dao.Client       `json:"client"`
	Role        dao.Role         `json:"role"`
	Branch      dao.ClientBranch `json:"branch"`
	Username    string           `bson:"username" json:"username"`
	Name        string           `bson:"name" json:"name"`
	Email       string           `bson:"email" json:"email"`
	PhoneNumber string           `bson:"phone_number" json:"phone_number"`
	Address     string           `bson:"address" json:"address"`
	IsCompany   bool             `bson:"is_company" json:"is_company"`
}

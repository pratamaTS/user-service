package dao

type ParentMenu struct {
	BaseModel   `bson:",inline"`
	Icon        string `bson:"icon" json:"icon"`
	Title       string `bson:"title" json:"title"`
	Description string `bson:"description" json:"description"`
	Sort        int    `bson:"sort" json:"sort"`
	IsActive    bool   `json:"is_active" bson:"is_active"`
}

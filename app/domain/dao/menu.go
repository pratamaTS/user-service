package dao

type Menu struct {
	BaseModel   `bson:",inline"`
	ParentUUID  string   `bson:"parent_uuid" json:"parent_uuid"`
	Icon        string   `bson:"icon" json:"icon"`
	Title       string   `bson:"title" json:"title"`
	Description string   `bson:"description" json:"description"`
	Href        string   `bson:"href" json:"href"`
	Owner       string   `bson:"owner" json:"owner"`
	Sort        int      `bson:"sort" json:"sort"`
	Function    []string `bson:"function" json:"function"`
	IsActive    bool     `json:"is_active" bson:"is_active"`
}

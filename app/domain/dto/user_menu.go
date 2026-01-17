package dto

type UserMenu struct {
	UUID           string   `json:"uuid"            bson:"uuid"`
	Name           string   `json:"name"            bson:"title"`
	Path           string   `json:"path"            bson:"href"`
	Icon           string   `json:"icon"            bson:"icon"`
	SortOrder      int      `json:"sort_order"      bson:"sort"`
	Parent         *Parent  `json:"parent"     bson:"parent"`
	AccessFunction []string `json:"access_function" bson:"-"`
}

type Parent struct {
	UUID      string `json:"uuid"       bson:"uuid"`
	Name      string `json:"name"       bson:"title"`
	Path      string `json:"path"       bson:"href"`
	Icon      string `json:"icon"       bson:"icon"`
	SortOrder int    `json:"sort_order" bson:"sort"`
}

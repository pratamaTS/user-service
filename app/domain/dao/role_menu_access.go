package dao

type RoleMenuAccess struct {
	BaseModel     `bson:",inline"`
	RoleUUID      string           `bson:"role_uuid" json:"role_uuid"`
	AccesibleMenu []AccessibleMenu `bson:"accessible_menus" json:"accessible_menus"`
}

type AccessibleMenu struct {
	MenuUUID       string   `bson:"menu_uuid" json:"menu_uuid"`
	AccessFunction []string `bson:"access_function" json:"access_function"`
}

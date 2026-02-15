package dao

type ProductUnit struct {
	Name             string  `bson:"name" json:"name"`
	ConversionToBase float64 `bson:"conversion_to_base" json:"conversion_to_base"`
}

type Product struct {
	BaseModel `bson:",inline"`

	BranchUUID  string        `bson:"branch_uuid" json:"branch_uuid"`
	SKU         string        `bson:"sku" json:"sku"`
	Barcode     string        `bson:"barcode" json:"barcode"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
	BaseUnit    string        `bson:"base_unit" json:"base_unit"`
	Units       []ProductUnit `bson:"units" json:"units"`
	Cost        float64       `bson:"cost" json:"cost"`
	Price       float64       `bson:"price" json:"price"`
	Image       string        `bson:"image" json:"image"`
	Stock       int64         `bson:"stock" json:"stock"`
	IsActive    bool          `bson:"is_active" json:"is_active"`
	CreatedBy   string        `bson:"created_by" json:"created_by"`
}

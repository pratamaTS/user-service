package dao

type ProductUnit struct {
	// contoh: "pcs", "box", "kg"
	Name string `bson:"name" json:"name"`

	// conversion_to_base: unit ini berapa "base_unit"
	// contoh base_unit=pcs, box=12 berarti conversion_to_base=12
	ConversionToBase float64 `bson:"conversion_to_base" json:"conversion_to_base"`
}

type Product struct {
	BaseModel `bson:",inline"`

	BranchUUID string `bson:"branch_uuid" json:"branch_uuid" validate:"required"`

	// identifier
	Image   string `bson:"image" json:"image"`
	SKU     string `bson:"sku" json:"sku"`
	Barcode string `bson:"barcode" json:"barcode"`

	Name        string `bson:"name" json:"name" validate:"required"`
	Description string `bson:"description" json:"description"`

	// units
	BaseUnit string        `bson:"base_unit" json:"base_unit" validate:"required"`
	Units    []ProductUnit `bson:"units" json:"units"`

	// pricing
	Cost  float64 `bson:"cost" json:"cost"`
	Price float64 `bson:"price" json:"price"`

	IsActive  bool   `bson:"is_active" json:"is_active"`
	CreatedBy string `bson:"created_by" json:"created_by"`
}

package dao

type BranchStock struct {
	BaseModel   `bson:",inline"`
	BranchUUID  string `bson:"branch_uuid" json:"branch_uuid"`
	ProductUUID string `bson:"product_uuid" json:"product_uuid"`
	QtyBase     int64  `bson:"qty_base" json:"qty_base"` // stok dalam base unit (pcs)
}

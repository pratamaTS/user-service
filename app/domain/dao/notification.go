package dao

import "time"

type Notification struct {
	UUID       string `bson:"uuid" json:"uuid"`
	ClientUUID string `bson:"client_uuid" json:"client_uuid"`

	// Targeting:
	// - BranchUUID untuk notif scope branch (gudang / tujuan)
	// - UserUUID untuk notif personal (driver terpilih)
	// Owner dapat semua notif via fetch by ClientUUID (tidak perlu disimpan khusus owner)
	BranchUUID string `bson:"branch_uuid" json:"branch_uuid"`
	UserUUID   string `bson:"user_uuid" json:"user_uuid"`

	Title   string `bson:"title" json:"title"`
	Message string `bson:"message" json:"message"`

	// optional UI hint
	Icon string `bson:"icon" json:"icon"` // success | warning | info
	Type string `bson:"type" json:"type"` // STOCK_TRANSFER | POS | STOCK
	Ref  string `bson:"ref" json:"ref"`   // uuid trx/transfer

	IsRead bool `bson:"is_read" json:"is_read"`

	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	CreatedAtStr string    `bson:"created_at_str" json:"created_at_str"`
}

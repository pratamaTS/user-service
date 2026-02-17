package dao

type AttendanceType string
type AttendanceStatus string

const (
	AttendanceCheckIn  AttendanceType = "CHECKIN"
	AttendanceCheckOut AttendanceType = "CHECKOUT"

	AttendancePresent AttendanceStatus = "PRESENT"
	AttendanceLate    AttendanceStatus = "LATE"
	AttendanceAbsent  AttendanceStatus = "ABSENT"
)

type AttendanceLocation struct {
	Latitude   float64 `bson:"latitude" json:"latitude"`
	Longitude  float64 `bson:"longitude" json:"longitude"`
	Address    string  `bson:"address" json:"address"`
	DistanceM  float64 `bson:"distance_m" json:"distance_m"`
	InRadius   bool    `bson:"in_radius" json:"in_radius"`
	MaxRadiusM int16   `bson:"max_radius_m" json:"max_radius_m"`
}

type Attendance struct {
	BaseModelV2 `bson:",inline"`

	ClientUUID string `bson:"client_uuid" json:"client_uuid"`
	BranchUUID string `bson:"branch_uuid" json:"branch_uuid"`

	UserUUID string `bson:"user_uuid" json:"user_uuid"`
	UserName string `bson:"user_name" json:"user_name"`

	Type   AttendanceType   `bson:"type" json:"type"`     // CHECKIN / CHECKOUT
	Status AttendanceStatus `bson:"status" json:"status"` // PRESENT/LATE/ABSENT

	Note string `bson:"note" json:"note"`

	Location AttendanceLocation `bson:"location" json:"location"`

	// untuk “open session”:
	IsOpen bool `bson:"is_open" json:"is_open"`
}

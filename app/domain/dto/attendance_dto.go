package dto

type AttendanceUpsertRequest struct {
	BranchUUID string `json:"branch_uuid"`
	Type       string `json:"type"`   // CHECKIN / CHECKOUT
	Status     string `json:"status"` // PRESENT/LATE/ABSENT
	Note       string `json:"note"`

	// akan diisi dari JWT (lebih aman), tapi boleh fallback dari body
	UserUUID string `json:"user_uuid"`
	UserName string `json:"user_name"`

	// lokasi dari frontend
	LocationLatitude  float64 `json:"location_latitude"`
	LocationLongitude float64 `json:"location_longitude"`
	LocationAddress   string  `json:"location_address"`
}

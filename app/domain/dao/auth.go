package dao

type Auth struct {
	BaseModel        `bson:",inline"`
	UserUUID         string `bson:"user_uuid" json:"user_uuid"`
	Subject          string `bson:"subject" json:"subject"`
	SessionID        string `bson:"session_id,omitempty" json:"session_id"`
	Token            string `bson:"token" json:"token"`
	LoginAt          int64  `bson:"login_at" json:"login_at"`
	RefreshToken     string `bson:"refresh_token" json:"refresh_token"`
	RefreshExpiredAt int64  `bson:"refresh_expired_at" json:"refresh_expired_at"`
	ExpiredAt        int64  `bson:"expired_at" json:"expired_at"`
	Status           string `bson:"status" json:"status"`
}

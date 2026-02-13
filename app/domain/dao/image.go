package dao

import "time"

type Image struct {
	UUID         string    `bson:"uuid" json:"uuid"`
	Bucket       string    `bson:"bucket" json:"bucket"`
	Key          string    `bson:"key" json:"key"`
	URL          string    `bson:"url" json:"url"`
	ContentType  string    `bson:"content_type" json:"content_type"`
	Size         int64     `bson:"size" json:"size"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	CreatedAtStr string    `bson:"created_at_str" json:"created_at_str"`
}

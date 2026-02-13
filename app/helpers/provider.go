package helpers

import "os"

func ProvideDBName() string {
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "user_service_db"
	}
	return dbName
}

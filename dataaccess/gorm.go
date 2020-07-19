package dataaccess

import (
	logger "github.com/sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

//Method to get local postgres database connection with GORM handle
func GetConnection() *gorm.DB {
	conn, err := gorm.Open(Configuration.DBConfig.Database, buildDBURI())
	if err != nil {
		panic("failed to connect database")
	}
	return conn
}

//Close an open gorm database connection
func CloseConnection(db *gorm.DB) {
	logger.Debug("Closing database connection", db)
	defer db.Close()
}

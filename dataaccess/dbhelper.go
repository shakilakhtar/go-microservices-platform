package dataaccess

import (
	"shakilakhtar/go-microservices-platform/utils"
	logger "github.com/sirupsen/logrus"
	"sync"
	"bytes"
)

type (
	configuration struct {
		DBConfig dbConfiguration
	}

	dbConfiguration struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Database string `json:"database"`
		Schema   string `json:"schema"`
		Username string `json:"username"`
		Password string `json:"password"`
		SSLMode  string `json:"sslmode"`
	}
)

const (
	DB_CONFIG_FILE = "dbconfig.json"
)

//A singleton context
var Configuration *configuration
var once sync.Once

// Load application configurations from config files and decode into context
func LoadDBConfigurationFromFile(location string) {
	GetConfiguration()
	if location == "" {
		logger.Info("location found empty trying with default db configuration %s", DB_CONFIG_FILE)
		location = "config"
	}
	//load database configurations from default config file
	utils.LoadConfig(location+"/"+DB_CONFIG_FILE, &Configuration.DBConfig)
	//else{
	//   //load database configurations
	//   loadConfig(configFile, &Configuration.DBConfig)
	// }

}

//Prepare Database connection URI from config details
func buildDBURI() string {
	var buffer bytes.Buffer
	buffer.WriteString("host=")
	buffer.WriteString(Configuration.DBConfig.Host)
	buffer.WriteString(" port=")
	buffer.WriteString(Configuration.DBConfig.Port)
	buffer.WriteString(" user=")
	buffer.WriteString(Configuration.DBConfig.Username)
	buffer.WriteString(" dbname=")
	buffer.WriteString(Configuration.DBConfig.Schema)
	buffer.WriteString(" sslmode=")
	buffer.WriteString(Configuration.DBConfig.SSLMode)
	buffer.WriteString(" password=")
	buffer.WriteString(Configuration.DBConfig.Password)

	logger.Info("Database connection URI", "DBURI", buffer.String())

	return buffer.String()
}

// GetInstance returns a singleton instance of configuration.
func GetConfiguration() *configuration {
	if Configuration == nil {
		once.Do(func() {
			Configuration = &configuration{}
		})
	}
	return Configuration

}

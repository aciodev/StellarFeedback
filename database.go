package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
)

var (
	Hostname = "localhost"
	Port     = 5432
	Username = "example"
	Password = "changeme"
	Database = "example"
)

// Config - The structural representation of the VIPER config file.
type Config struct {
	Hostname string `mapstructure:"HOSTNAME"`
	Port     int    `mapstructure:"PORT"`
	Username string `mapstructure:"USERNAME"`
	Password string `mapstructure:"PASSWORD"`
	Database string `mapstructure:"DATABASE"`
	HttpPort string `mapstructure:"HTTP_PORT"`
}

// LoadConfig - Loads the config at a given path, returning the
// unmarshalled data as the return type.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

// ConnectPostgres - Connects to the Postgresql server.
func ConnectPostgres() *sql.DB {
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		Hostname, Port, Username, Password, Database)

	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Println(err)
		return nil
	}

	return db
}

// LogFeedback - Logs the feedback for a given BugReport.
func LogFeedback(report BugReport) bool {
	db := ConnectPostgres()
	if db == nil {
		log.Println("Cannot connect to PostgreSQL!")
		_ = db.Close()
		return false
	}

	defer db.Close()

	_, err := db.Query("INSERT INTO log (title, body, link, version) VALUES($1, $2, $3, $4)\n", report.Title, report.Body, report.Link, report.Version)
	if err != nil {
		log.Println("Query:", err)
		return false
	}

	return true
}

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
	Token    string `mapstructure:"TOKEN"`
}

// DatabaseReport - The structural representation of the SQL table.
type DatabaseReport struct {
	Id        int
	CreatedAt string
	Title     string
	Body      string
	Link      string
	Version   string
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

// LogFeedback - Logs the feedback for a given UserFeedback.
func LogFeedback(report UserFeedback) bool {
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

// GetRecentFeedback - Gets the recent feedback from the database.
func GetRecentFeedback() []DatabaseReport {
	db := ConnectPostgres()
	if db == nil {
		log.Println("Cannot connect to PostgreSQL!")
		_ = db.Close()
		return []DatabaseReport{}
	}

	defer db.Close()

	rows, err := db.Query("SELECT id,created_at,title,body,link,version FROM log ORDER BY id desc LIMIT 10")

	if err != nil {
		log.Println("Query:", err)
		return []DatabaseReport{}
	}

	defer rows.Close()
	reports := reportsFromRow(rows)
	return reports
}

// GetFeedbackBefore - Gets feedback from the database before a given id.
func GetFeedbackBefore(id int) []DatabaseReport {
	db := ConnectPostgres()
	if db == nil {
		log.Println("Cannot connect to PostgreSQL!")
		_ = db.Close()
		return []DatabaseReport{}
	}

	defer db.Close()

	rows, err := db.Query("SELECT id,created_at,title,body,link,version FROM log WHERE id < $1 ORDER BY id desc LIMIT 10", id)

	if err != nil {
		log.Println("Query:", err)
		return []DatabaseReport{}
	}

	defer rows.Close()
	reports := reportsFromRow(rows)
	return reports
}

// reportsFromRow - Private helper method to scan sql.Row objects into
// structs. This data is then marshalled into JSON for the request.
func reportsFromRow(rows *sql.Rows) []DatabaseReport {
	reports := make([]DatabaseReport, 0)

	var id int
	var createdAt string
	var title string
	var body string
	var link string
	var version string

	for rows.Next() {
		err := rows.Scan(&id, &createdAt, &title, &body, &link, &version)
		if err != nil {
			log.Println(err)
			continue
		}

		var report = DatabaseReport{id, createdAt, title, body, link, version}
		reports = append(reports, report)
	}

	return reports
}

package db

import (
	"database/sql"

	// Postgres database driver
	_ "github.com/lib/pq"

	"fmt"
)

//ReportTable : global name for table with reports
var ReportTable string

//FragmentTable : global name for table with fragments
var FragmentTable string

//RuleTable : global name for table with rules
var RuleTable string

//Init :  init checks & table creation
func Init(conn *sql.DB) (err error) {
	FragmentTable = "fragments"
	ReportTable = "reports"
	RuleTable = "rules"

	exist, err := CheckExists(FragmentTable, conn)
	if err != nil {
		return
	}
	if !exist {
		if err = createFragmentTable(FragmentTable, conn); err != nil {
			return err
		}
	}

	exist, err = CheckExists(ReportTable, conn)
	if err != nil {
		return
	}
	if !exist {
		if err = createReportTable(ReportTable, conn); err != nil {
			return err
		}
	}

	exist, err = CheckExists(RuleTable, conn)
	if err != nil {
		return
	}
	if !exist {
		if err = createRulesTable(RuleTable, conn); err != nil {
			return err
		}
	}

	return
}

// Connect to database
func Connect(name, password, database string) (db *sql.DB, err error) {
	ConnectURI := fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", name, password, database)
	db, err = sql.Open("postgres", ConnectURI)
	return
}

//CheckExists : Check if table exists in db
func CheckExists(tableName string, conn *sql.DB) (exist bool, err error) {
	query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema=$2 AND table_name=$1);"
	schema := "public"
	row := conn.QueryRow(query, tableName, schema)

	err = row.Scan(&exist)
	return
}

//DropTable : drops table if exists
func DropTable(tableName string, conn *sql.DB) (err error) {
	query := "DROP TABLE IF EXISTS " + tableName + " CASCADE;"
	_, err = conn.Exec(query)
	return
}

func createFragmentTable(tableName string, conn *sql.DB) (err error) {
	query := "CREATE TABLE " + tableName + " (id serial, content bytea, reject_id integer, report_id integer, shahash bigint, keywords jsonb);"
	_, err = conn.Exec(query)
	return
}

func createReportTable(tableName string, conn *sql.DB) (err error) {
	query := "CREATE TABLE " + tableName + " (id serial, shahash bigint, status varchar, data bytea, time integer);"
	_, err = conn.Exec(query)
	return
}

func createRulesTable(tableName string, conn *sql.DB) (err error) {
	query := "CREATE TABLE " + tableName + " (id serial, name varchar, rule varchar);"
	_, err = conn.Exec(query)
	if err != nil {
		return
	}

	query = "INSERT INTO " + tableName + " (name, rule) VALUES ($1, $2)"
	_, err = conn.Exec(query, "manual", "")
	if err != nil {
		return
	}

	_, err = conn.Exec(query, "verified", "")
	if err != nil {
		return
	}

	_, err = conn.Exec(query, "auto_remove", "")
	if err != nil {
		return
	}
	return
}

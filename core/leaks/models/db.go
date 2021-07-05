package models

import (
	"database/sql"
	"encoding/json"
	"regexp"

	"fmt"

	//Postgresql driver
	_ "github.com/lib/pq"
	"github.com/megamon/core/utils"
)

//Manager : database manager for all types
type Manager struct {
	Database *sql.DB
}

//ReportTable : global name for table with reports
var ReportTable = "reports"

//FragmentTable : global name for table with fragments
var FragmentTable = "fragments"

//RuleTable : global name for table with rules
var RuleTable = "rules"

//KeywordsTable : global name for table with keywords
var KeywordsTable = "keywords"

//Init : Manager constructor
func (manager *Manager) Init() (err error) {
	creds := utils.Settings.DBCredentials
	conn, err := Connect(creds.Name, creds.Password, creds.DBHostName, creds.Database)
	if err != nil {
		return
	}

	manager.Database = conn
	return
}

//Close : Manager destructor
func (manager *Manager) Close() {
	if manager.Database != nil {
		manager.Database.Close()
	}
	return
}

//InsertTextFragment : insert text fragment into db
func (manager *Manager) InsertTextFragment(frag *TextFragment) (ID int, err error) {
	query := "INSERT INTO " + FragmentTable + " (content, reject_id, report_id, type, shahash, keywords) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;"
	kwData, err := json.Marshal(frag.Keywords)
	content := []byte(frag.Text)

	if err != nil {
		return 0, err
	}

	err = manager.Database.QueryRow(query, content, frag.RejectID, frag.ReportID, frag.Type, frag.ShaHash, kwData).Scan(&ID)
	return
}

//UpdateTextFragmentRejectID : updates text fragment in db
func (manager *Manager) UpdateTextFragmentRejectID(ID, rejectID int) (err error) {
	query := "UPDATE " + FragmentTable + " SET reject_id=$2 WHERE id=$1;"
	_, err = manager.Database.Exec(query, ID, rejectID)
	return
}

//DeleteTextFragmentByID : deletes text fragment from db
func (manager *Manager) DeleteTextFragmentByID(ID int) (err error) {
	query := "DELETE FROM " + FragmentTable + " WHERE id=$1;"
	_, err = manager.Database.Exec(query, ID)
	return
}

//CountTextFragments : return count of text fragments with defined type
func (manager *Manager) CountTextFragments(field string, value int, extensions ...string) (count int, err error) {
	extension := ""
	for _, ext := range extensions {
		extension += ext
	}

	query := "SELECT COUNT(id) FROM " + FragmentTable + " WHERE " + field + "=$1 " + extension + ";"
	row := manager.Database.QueryRow(query, value)
	err = row.Scan(&count)
	return
}

//SelectTextFragment : select text fragment from db
func (manager *Manager) SelectTextFragment(field string, value int, extensions ...string) (frags []TextFragment, err error) {
	extension := ""
	for _, ext := range extensions {
		extension += ext
	}

	query := "SELECT id, content, reject_id, report_id, type, shahash, keywords FROM " + FragmentTable + " WHERE " + field + "=$1 " + extension + ";"
	rows, err := manager.Database.Query(query, value)

	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var frag TextFragment
		var content []byte
		var kwData []byte

		err = rows.Scan(&frag.ID, &content, &frag.RejectID, &frag.ReportID, &frag.Type, &frag.ShaHash, &kwData)
		if err != nil {
			return
		}

		err = json.Unmarshal(kwData, &frag.Keywords)
		if err != nil {
			return
		}

		frag.Text = string(content)
		frags = append(frags, frag)
	}
	return
}

//CheckTextFragmentDuplicate : Check for text fragment with the same hash
func (manager *Manager) CheckTextFragmentDuplicate(ShaHash string) (exist bool, err error) {
	query := "SELECT EXISTS(SELECT id FROM " + FragmentTable + " WHERE shahash=$1);"

	row := manager.Database.QueryRow(query, ShaHash)
	err = row.Scan(&exist)
	return
}

//InsertReport : inser report to db
func (manager *Manager) InsertReport(report Report) (ID int, err error) {
	query := "INSERT INTO " + ReportTable + " (shahash, type, status, data, time) VALUES ($1, $2, $3, $4, $5) RETURNING id;"

	err = manager.Database.QueryRow(query, report.ShaHash, report.Type, report.Status, report.Data, report.Time).Scan(&ID)
	return
}

//UpdateReport : update report in db
func (manager *Manager) UpdateReport(field, queryField string, newValue interface{}, queryValue interface{}) (err error) {
	query := "UPDATE " + ReportTable + " SET $1=$2 WHERE $3=$4;"
	_, err = manager.Database.Exec(query, field, newValue, queryField, queryValue)
	return
}

//UpdateReportStatus : updates status of the report
func (manager *Manager) UpdateReportStatus(reportID int, status string) (err error) {
	query := "UPDATE " + ReportTable + " SET status=$2 WHERE id=$1;"
	_, err = manager.Database.Exec(query, reportID, status)
	return
}

//UpdateReportTime : updates timestamp
func (manager *Manager) UpdateReportTime(reportID int, timestamp int) (err error) {
	query := "UPDATE " + ReportTable + " SET time=$2 WHERE id=$1;"
	_, err = manager.Database.Exec(query, reportID, timestamp)
	return
}

//DeleteReportByID : delete reprort from db
func (manager *Manager) DeleteReportByID(ID int) (err error) {
	query := "DELETE FROM " + ReportTable + " WHERE id=$1;"
	_, err = manager.Database.Exec(query, ID)
	return
}

//SelectReportByID : select report from db
func (manager *Manager) SelectReportByID(ID int) (rep Report, err error) {
	var shaHashStr string

	query := "SELECT id, type, status, data, shahash, time FROM " + ReportTable + " WHERE id=$1;"
	row := manager.Database.QueryRow(query, ID)
	err = row.Scan(&rep.ID, &rep.Type, &rep.Status, &rep.Data, &shaHashStr, &rep.Time)
	return
}

//SelectReportTypes : select report types
func (manager *Manager) SelectReportTypes() (types []string, err error) {
	query := "SELECT DISTINCT type FROM " + ReportTable + ";"
	rows, err := manager.Database.Query(query)
	if err != nil {
		return
	}

	types = make([]string, 0, 16)
	defer rows.Close()
	for rows.Next() {
		var reportType string
		err = rows.Scan(&reportType)
		if err != nil {
			return
		}
		types = append(types, reportType)
	}
	return
}

//SelectReportByStatus : select report by it's type & status
func (manager *Manager) SelectReportByStatus(reportType, status string, extensions ...string) (reports []Report, err error) {
	extension := ""
	for _, ext := range extensions {
		extension += ext
	}

	query := "SELECT id, type, status, data, shahash, time FROM " + ReportTable + " WHERE type=$1 AND status=$2 " + extension + ";"
	rows, err := manager.Database.Query(query, reportType, status)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var rep Report

		err = rows.Scan(&rep.ID, &rep.Type, &rep.Status, &rep.Data, &rep.ShaHash, &rep.Time)
		if err != nil {
			return
		}

		reports = append(reports, rep)
	}

	return
}

//CheckReportDuplicate : Check for report with the same hash
func (manager *Manager) CheckReportDuplicate(ShaHash string) (exist bool, err error) {
	query := "SELECT EXISTS(SELECT id FROM " + ReportTable + " WHERE shahash=$1);"

	row := manager.Database.QueryRow(query, ShaHash)
	err = row.Scan(&exist)
	return
}

//CountReports : return count of text fragments with defined type
func (manager *Manager) CountReports(reportType string, extensions ...string) (count int, err error) {
	extension := ""
	for _, ext := range extensions {
		extension += ext
	}

	query := "SELECT COUNT(id) FROM " + ReportTable + " WHERE type=$1" + extension + ";"
	row := manager.Database.QueryRow(query)
	err = row.Scan(&count)
	return
}

//InsertRule : inser rule into db
func (manager *Manager) InsertRule(rule RejectRule) (ID int, err error) {
	//Verify data consistency
	if rule.Expr == nil {
		rule.Expr, err = regexp.Compile(rule.Rule)
		if err != nil {
			return
		}
	}

	query := "INSERT INTO " + RuleTable + " (name, rule) VALUES ($1, $2) RETURNING id;"
	err = manager.Database.QueryRow(query, rule.Name, rule.Rule).Scan(&ID)
	return
}

//UpdateRule : update rule in database
func (manager *Manager) UpdateRule(rule RejectRule) (err error) {
	err = fmt.Errorf("Not implemented")
	return
}

//DeleteRuleByID : update rule in database
func (manager *Manager) DeleteRuleByID(ID int) (err error) {
	query := "DELETE FROM " + RuleTable + " WHERE id=$1;"
	_, err = manager.Database.Exec(query, ID)
	return
}

//SelectRuleByID : select rejection rule by id
func (manager *Manager) SelectRuleByID(ID int) (rule RejectRule, err error) {
	query := "SELECT id, name, rule FROM " + RuleTable + " WHERE id=$1;"
	row := manager.Database.QueryRow(query, ID)
	err = row.Scan(&rule.ID, &rule.Name, &rule.Rule)
	if err != nil {
		return
	}

	rule.Expr, err = regexp.Compile(rule.Rule)
	if err != nil {
		return
	}

	return
}

//SelectAllRules : select all rejection rules from database
func (manager *Manager) SelectAllRules() (rules []RejectRule, err error) {
	query := "SELECT id, name, rule FROM " + RuleTable + ";"
	rows, err := manager.Database.Query(query)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var rule RejectRule
		err = rows.Scan(&rule.ID, &rule.Name, &rule.Rule)
		if err != nil {
			return
		}

		rule.Expr, err = regexp.Compile(rule.Rule)
		if err != nil {
			return
		}

		rules = append(rules, rule)
	}
	return
}

//InsertKeyword : insert keyword to the databese
func (manager *Manager) InsertKeyword(keyword string, wordType int) (ID int, err error) {
	query := "INSERT INTO " + KeywordsTable + " (keyword, type)  VALUES  ($1, $2) RETURNING id;"
	err = manager.Database.QueryRow(query, keyword, wordType).Scan(&ID)
	return
}

//DeleteKeyword : delete keyword from database
func (manager *Manager) DeleteKeyword(ID int) (err error) {
	query := "DELETE FROM " + KeywordsTable + " WHERE id=$1;"
	_, err = manager.Database.Exec(query, ID)
	return
}

//SelectKeywordByType : select all keywords with the same type
func (manager *Manager) SelectKeywordByType(wordType int) (keywords []Keyword, err error) {
	query := "SELECT id, keyword, type FROM " + KeywordsTable + " WHERE type=$1;"
	rows, err := manager.Database.Query(query, wordType)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var keyword Keyword
		err = rows.Scan(&keyword.ID, &keyword.Value, &keyword.Type)
		if err != nil {
			return
		}
		keywords = append(keywords, keyword)
	}
	return
}

//SelectAllKeywords : select all keywords from database
func (manager *Manager) SelectAllKeywords() (keywords []Keyword, err error) {
	query := "SELECT id, keyword, type FROM " + KeywordsTable + ";"
	rows, err := manager.Database.Query(query)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var keyword Keyword
		err = rows.Scan(&keyword.ID, &keyword.Value, &keyword.Type)
		if err != nil {
			return
		}
		keywords = append(keywords, keyword)
	}
	return
}

//SelectKeywordByID : select particular keyword from database by its id
func (manager *Manager) SelectKeywordByID(ID int) (keyword Keyword, err error) {
	query := "SELECT id, keyword, type FROM " + KeywordsTable + " WHERE id=$1"
	row := manager.Database.QueryRow(query, ID)
	err = row.Scan(&keyword.ID, &keyword.Value, &keyword.Type)
	return
}

//Init :  init checks & table creation
func Init(conn *sql.DB) (err error) {
	tables := make(map[string](func(name string, conn *sql.DB) (err error)), 10)

	tables[FragmentTable] = createFragmentTable
	tables[ReportTable] = createReportTable
	tables[RuleTable] = createRulesTable
	tables[KeywordsTable] = createKeywordsTable

	for table := range tables {
		exist, err := CheckExists(table, conn)
		if err != nil {
			return err
		}
		if !exist {
			err = tables[table](table, conn)
			if err != nil {
				return err
			}
		}
	}

	return
}

// Connect to database
func Connect(name, password, hostname, database string) (db *sql.DB, err error) {
	ConnectURI := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", name, password, hostname, database)
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
	query := "CREATE TABLE " + tableName + " (id serial, content bytea, reject_id integer, report_id integer,type varchar, shahash varchar PRIMARY KEY, keywords jsonb);"
	_, err = conn.Exec(query)
	return
}

func createReportTable(tableName string, conn *sql.DB) (err error) {
	query := "CREATE TABLE " + tableName + " (id serial, shahash varchar PRIMARY KEY, status varchar, type varchar, data bytea, time integer);"
	_, err = conn.Exec(query)
	return
}

func createKeywordsTable(tableName string, conn *sql.DB) (err error) {
	query := "CREATE TABLE " + tableName + " (id serial, type int, keyword varchar);"
	_, err = conn.Exec(query)
	return
}

func createRulesTable(tableName string, conn *sql.DB) (err error) {
	query := "CREATE TABLE " + tableName + " (id serial, name varchar, rule varchar) ;"
	_, err = conn.Exec(query)
	if err != nil {
		return
	}

	query = "INSERT INTO " + tableName + " (name, rule) VALUES ($1, $2)"
	predefined := []string{"none", "manual", "verified", "auto_removed"}
	for _, rule := range predefined {
		_, err = conn.Exec(query, rule, "")
		if err != nil {
			return
		}
	}

	return
}

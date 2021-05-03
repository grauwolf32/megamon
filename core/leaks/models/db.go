package models

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"

	"fmt"

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
	conn, err := Connect(creds.Name, creds.Password, creds.Database)
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
	query := "INSERT INTO " + FragmentTable + " (content, reject_id, report_id, shahash, keywords) VALUES ($1, $2, $3, $4, $5) RETURNING id;"
	kwData, err := json.Marshal(frag.Keywords)
	shaHash := fmt.Sprintf("%x", string(frag.ShaHash[:]))
	content := []byte(frag.Text)

	if err != nil {
		return 0, err
	}

	err = manager.Database.QueryRow(query, content, frag.RejectID, frag.ReportID, shaHash, kwData).Scan(&ID)
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

//SelectTextFragment : select text fragment from db
func (manager *Manager) SelectTextFragment(field string, value int) (frags []TextFragment, err error) {
	query := "SELECT id, content, reject_id, report_id, shahash, keywords FROM " + FragmentTable + " WHERE " + field + "=$1;"
	rows, err := manager.Database.Query(query, value)

	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var frag TextFragment
		var content []byte
		var kwData []byte
		var shaHashStr string

		err = rows.Scan(&frag.ID, &content, &frag.RejectID, &frag.ReportID, &shaHashStr, &kwData)
		if err != nil {
			return
		}

		err = json.Unmarshal(kwData, &frag.Keywords)
		if err != nil {
			return
		}

		shaHash, DecErr := hex.DecodeString(shaHashStr)
		if DecErr != nil {
			return
		}

		copy(frag.ShaHash[:], shaHash[:20])
		frag.Text = string(content)
		frags = append(frags, frag)
	}
	return
}

//InsertReport : inser report to db
func (manager *Manager) InsertReport(report Report) (ID int, err error) {
	query := "INSERT INTO " + ReportTable + " (type, status, data, time) VALUES ($1, $2, $3, $4) RETURNING id;"
	shaHash := fmt.Sprintf("%x", string(report.ShaHash[:]))
	fmt.Printf("%s", shaHash)

	err = manager.Database.QueryRow(query, shaHash, report.Status, report.Data, report.Time).Scan(&ID)
	return
}

//UpdateReport : update report in db
func (manager *Manager) UpdateReport(field, queryField string, newValue interface{}, queryValue interface{}) (err error) {
	query := "UPDATE " + ReportTable + " SET $1=$2 WHERE $3=$4;"
	_, err = manager.Database.Exec(query, field, newValue, queryField, queryValue)
	return
}

//DeleteReportByID : delete reprort from db
func (manager *Manager) DeleteReportByID(ID int) (err error) {
	query := "DELETE FROM " + ReportTable + " WHERE id=$1;"
	_, err = manager.Database.Exec(query, ID)
	return
}

//SelectReportByID : select report from db
func (manager *Manager) SelectReportByID(ID int) (reports []Report, err error) {
	query := "SELECT id, type, status, data, shahash, time FROM " + ReportTable + " WHERE id=$1;"

	rows, err := manager.Database.Query(query, ID)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var rep Report
		var shaHashStr string

		err = rows.Scan(&rep.ID, &rep.Type, &rep.Status, &rep.Data, &shaHashStr, &rep.Time)
		if err != nil {
			return
		}

		shaHash, DecErr := hex.DecodeString(shaHashStr)
		if DecErr != nil {
			return
		}

		copy(rep.ShaHash[:], shaHash[:20])
		reports = append(reports, rep)
	}

	return
}

//SelectReportByStatus : select report by it's type & status
func (manager *Manager) SelectReportByStatus(reportType, status string) (reports []Report, err error) {
	query := "SELECT id, type, status, data, shahash, time FROM " + ReportTable + " WHERE type=$1 AND status=$2;"
	rows, err := manager.Database.Query(query, reportType, status)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var rep Report
		var shaHashStr string

		err = rows.Scan(&rep.ID, &rep.Type, &rep.Status, &rep.Data, &shaHashStr, &rep.Time)
		if err != nil {
			return
		}

		shaHash, DecErr := hex.DecodeString(shaHashStr)
		if DecErr != nil {
			return
		}

		copy(rep.ShaHash[:], shaHash[:20])
		reports = append(reports, rep)
	}

	return
}

//CheckReportDuplicate : Check for report with the same hash
func (manager *Manager) CheckReportDuplicate(ShaHash []byte) (exist bool, err error) {
	query := "SELECT EXIST(SELECT id FROM " + ReportTable + " WHERE shahash=$1);"
	shaHash := fmt.Sprintf("%x", ShaHash)

	row := manager.Database.QueryRow(query, shaHash)
	err = row.Scan(&exist)
	return
}

//InsertRule : inser rule into db
func (manager *Manager) InsertRule(rule RejectRule) (ID int, err error) {
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
	query := "SELECT id, keyword, type FROM " + KeywordsTable + " WHERE type=$1"
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
	query := "CREATE TABLE " + tableName + " (id serial, content bytea, reject_id integer, report_id integer, shahash varchar, keywords jsonb);"
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

package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/megamon/core/leaks/helpers"
)

//Manager : database manager for all types
type Manager struct {
	Database *sql.DB
}

//Init : Manager constructor
func (manager *Manager) Init(conn *sql.DB) {
	manager.Database = conn
	return
}

//InsertTextFragment : insert text fragment into db
func (manager *Manager) InsertTextFragment(frag *helpers.TextFragment) (ID int, err error) {
	query := "INSERT INTO " + FragmentTable + " (content, reject_id, report_id, shahash, keywords) VALUES ($1, $2, $3, $4, $5) RETURNING id;"
	kwData, err := json.Marshal(frag.Keywords)
	content := []byte(frag.Text)

	if err != nil {
		return 0, err
	}

	err = manager.Database.QueryRow(query, content, frag.RejectID, frag.ReportID, frag.ShaHash, kwData).Scan(&ID)
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
func (manager *Manager) SelectTextFragment(field string, value int) (frags []helpers.TextFragment, err error) {
	query := "SELECT id, content, reject_id, report_id, shahash, keywords FROM " + FragmentTable + " WHERE " + field + "=$1;"
	rows, err := manager.Database.Query(query, value)

	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var frag helpers.TextFragment
		var content []byte
		var kwData []byte

		err = rows.Scan(&frag.ID, &content, &frag.RejectID, &frag.ReportID, &frag.ShaHash, &kwData)
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

//InsertReport : inser report to db
func (manager *Manager) InsertReport(report helpers.Report) (ID int, err error) {
	query := "INSERT INTO " + ReportTable + " (type, status, data, time) VALUES ($1, $2, $3, $4) RETURNING id;"
	err = manager.Database.QueryRow(query, report.ShaHash, report.Status, report.Data, report.Time).Scan(&ID)
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
func (manager *Manager) SelectReportByID(ID int) (reports []helpers.Report, err error) {
	query := "SELECT id, type, status, data, time FROM " + ReportTable + " WHERE id=$1;"

	rows, err := manager.Database.Query(query, ID)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var rep helpers.Report
		err = rows.Scan(&rep.ID, &rep.Type, &rep.Status, &rep.Data, &rep.Time)
		if err != nil {
			return
		}
	}

	return
}

//SelectReportByStatus : select report by it's type & status
func (manager *Manager) SelectReportByStatus(reportType, status string) (err error) {
	query := "SELECT id, type, status, data, time FROM " + ReportTable + " WHERE type=$1 AND status=$2;"
	rows, err := manager.Database.Query(query, reportType, status)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var rep helpers.Report
		err = rows.Scan(&rep.ID, &rep.Type, &rep.Status, &rep.Data, &rep.Time)
		if err != nil {
			return
		}
	}

	return
}

//InsertRule : inser rule into db
func (manager *Manager) InsertRule(rule helpers.RejectRule) (ID int, err error) {
	query := "INSERT INTO " + RuleTable + " (name, rule) VALUES ($1, $2) RETURNING id;"
	err = manager.Database.QueryRow(query, rule.Name, rule.Rule).Scan(&ID)
	return
}

//UpdateRule : update rule in database
func (manager *Manager) UpdateRule(rule helpers.RejectRule) (err error) {
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
func (manager *Manager) SelectRuleByID(ID int) (rule helpers.RejectRule, err error) {
	query := "SELECT id, name, rule FROM " + RuleTable + " WHERE id=$1;"
	row := manager.Database.QueryRow(query, ID)
	err = row.Scan(&rule.ID, &rule.Name, &rule.Rule)
	return
}

//SelectAllRules : select all rejection rules from database
func (manager *Manager) SelectAllRules() (rules []helpers.RejectRule, err error) {
	query := "SELECT id, name, rule FROM " + RuleTable + ";"
	rows, err := manager.Database.Query(query)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var rule helpers.RejectRule
		err = rows.Scan(&rule.ID, &rule.Name, &rule.Rule)
		if err != nil {
			return
		}
	}
	return
}

package models

import (
	"fmt"
	"os"
	"testing"

	"github.com/megamon/core/utils"
)

var testFragment string
var testReport string
var testRules string

func setup() {
	utils.InitConfig("../../../config/config.yaml")
	creds := utils.Settings.DBCredentials

	conn, err := Connect(creds.Name, creds.Password, creds.Database)
	defer conn.Close()

	FragmentTable = "fragment_test"
	ReportTable = "report_test"
	RuleTable = "rules_test"
	KeywordsTable = "keywords_test"

	if err != nil {
		panic(err)
	}

	err = Init(conn)
	if err != nil {
		panic(err)
	}
	return
}

func clean() {
	creds := utils.Settings.DBCredentials
	conn, err := Connect(creds.Name, creds.Password, creds.Database)
	defer conn.Close()

	tables := []string{FragmentTable, ReportTable, RuleTable, KeywordsTable}
	for _, table := range tables {
		if err = DropTable(table, conn); err != nil {
			panic(err)
		}
	}

	return
}

func TestMain(m *testing.M) {
	setup()
	retCode := m.Run()
	clean()
	os.Exit(retCode)
}

func TestConnect(t *testing.T) {
	creds := utils.Settings.DBCredentials
	conn, err := Connect(creds.Name, creds.Password, creds.Database)

	if err != nil {
		t.Errorf("%s", err.Error())
		return
	}

	err = conn.Ping()

	if err != nil {
		t.Errorf("%s", err.Error())
	}
	return
}

func TestTextFragmentOps1(t *testing.T) {
	var manager Manager
	manager.Init()
	defer manager.Close()

	var tf TextFragment
	tf.ReportID = 111
	tf.RejectID = 1
	tf.Text = "test fragment"
	tf.Type = "test"
	tf.ShaHash = fmt.Sprintf("%x", [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
	tf.Keywords = [][]int{{1, 2}, {3, 4}}

	ID, err := manager.InsertTextFragment(&tf)

	if err != nil {
		t.Errorf("%s", err.Error())
		return
	}

	tfs, err := manager.SelectTextFragment("id", ID)

	if err != nil {
		t.Errorf("%s", err.Error())
		return
	}

	if len(tfs) != 1 {
		t.Errorf("Expected length of fragments: 1 got %d", len(tfs))
		return
	}

	ntf := tfs[0]
	v := ntf.ReportID != tf.ReportID
	v = v || ntf.RejectID != tf.RejectID
	v = v || ntf.Text != tf.Text
	v = v || ntf.ShaHash != tf.ShaHash
	v = v || ntf.Type != tf.Type

	if v {
		t.Errorf("Input object: %v\n Aquired object: %v", tf, ntf)
	}
	return
}

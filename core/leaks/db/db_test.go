package db

import (
	"os"
	"testing"

	"../helpers"

	"../../config"
	"../../utils"
)

var testFragment string
var testReport string
var testRules string

func setup() {
	utils.InitConfig("../../../config/config.json")
	creds := config.Settings.DBCredentials

	conn, err := Connect(creds.Name, creds.Password, creds.Database)
	defer conn.Close()

	FragmentTable = "fragment_test"
	ReportTable = "report_test"
	RuleTable = "rules_test"

	if err != nil {
		panic(err)
	}

	exist, err := CheckExists(FragmentTable, conn)
	if err != nil {
		panic(err)
	}
	if !exist {
		if err = createFragmentTable(FragmentTable, conn); err != nil {
			panic(err)
		}
	}

	exist, err = CheckExists(ReportTable, conn)
	if err != nil {
		panic(err)
	}
	if !exist {
		if err = createReportTable(ReportTable, conn); err != nil {
			panic(err)
		}
	}

	exist, err = CheckExists(RuleTable, conn)
	if err != nil {
		panic(err)
	}
	if !exist {
		if err = createRulesTable(RuleTable, conn); err != nil {
			panic(err)
		}
	}
}

func clean() {
	creds := config.Settings.DBCredentials
	conn, err := Connect(creds.Name, creds.Password, creds.Database)
	defer conn.Close()

	if err = DropTable(FragmentTable, conn); err != nil {
		panic(err)
	}

	if err = DropTable(ReportTable, conn); err != nil {
		panic(err)
	}

	if err = DropTable(RuleTable, conn); err != nil {
		panic(err)
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
	creds := config.Settings.DBCredentials
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
	creds := config.Settings.DBCredentials
	conn, err := Connect(creds.Name, creds.Password, creds.Database)

	var manager Manager
	manager.Init(conn)

	var tf helpers.TextFragment
	tf.ReportID = 111
	tf.RejectID = 1
	tf.Text = "test fragment"
	tf.ShaHash = 1234
	tf.Keywords = [][]int{[]int{1, 2}, []int{3, 4}}

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

	if v {
		t.Errorf("Input object: %v\n Aquired object: %v", tf, ntf)
	}
	return
}

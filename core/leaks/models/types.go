package models

import "regexp"

//Report : report structure
type Report struct {
	ShaHash string `json:"sha1"`
	Time    int64  `json:"time"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Data    []byte `json:"data"`
	ID      int    `json:"id"`
}

//RejectRule : description of reject rule
type RejectRule struct {
	ID   int    `json:"id"`
	Rule string `json:"rule"`
	Name string `json:"name"`
	Expr *regexp.Regexp
}

//TextFragment : fragments of text with keywords
type TextFragment struct {
	ShaHash  string  `json:"sha1"`
	Text     string  `json:"text"`
	Type     string  `json:"type"`
	ID       int     `json:"id"`
	ReportID int     `json:"report_id"`
	RejectID int     `json:"reject_id"`
	Keywords [][]int `json:"keywords"`
}

//Keyword : auxilary data type
type Keyword struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
	Type  int    `json:"type"`
}

const (
	//KWSEARCHABLE : searchable keyword type
	KWSEARCHABLE = iota

	//KWINNER : non searchable keword type
	KWINNER
)

const (
	//RULENONE : default rejection status - not rejected
	RULENONE = iota

	//RULEMANUAL : fragment was manually rejected
	RULEMANUAL

	//RULEVERIFIED : fragment was manually verified
	RULEVERIFIED

	//RULEAUTOREMOVED : fragment was automatically removed by regexp
	RULEAUTOREMOVED
)

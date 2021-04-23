package helpers

//Report : report structure
type Report struct {
	ID      int      `json:"id"`
	Type    string   `json:"type"`
	Status  string   `json:"status"`
	ShaHash [20]byte `json:"sha1"`
	Data    []byte   `json:"data"`
	Time    int64    `json:"time"`
}

//RejectRule : description of reject rule
type RejectRule struct {
	ID   int    `json:"id"`
	Rule string `json:"rule"`
	Name string `json:"name"`
}

//TextFragment : fragments of text with keywords
type TextFragment struct {
	ID       int      `json:"id"`
	ReportID int      `json:"report_id"`
	RejectID int      `json:"reject_id"`
	Text     string   `json:"text"`
	ShaHash  [20]byte `json:"sha1"`
	Keywords [][]int  `json:"keywords"`
}

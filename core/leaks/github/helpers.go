package github

import (
	"fmt"

	"github.com/megamon/core/utils"
)

const (
	//MAXRESPONSEITEMS : max items in search response
	MAXRESPONSEITEMS = 100

	//MAXOFFSET : maximum offset supported by github API
	MAXOFFSET = 10
)

type gitRepoOwner struct {
	Login string `json:"login"`
	URL   string `json:"url"`
}

type gitRepo struct {
	Name     string       `json:"name"`
	FullName string       `json:"full_name"`
	Owner    gitRepoOwner `json:"owner"`
}

type gitRequestParams struct {
	Query   string
	Keyword string
	Offset  int
}

//GitSearchItem : search item format
type GitSearchItem struct {
	Name    string  `json:"name"`
	Path    string  `json:"path"`
	ShaHash string  `json:"sha"`
	URL     string  `json:"url"`
	GitURL  string  `json:"git_url"`
	HTMLURL string  `json:"html_url"`
	Repo    gitRepo `json:"repository"`
	Score   float32 `json:"score"`
}

//GitFetchItem : fetch response format
type GitFetchItem struct {
	Content  []byte `json:"content"`
	Encoding string `json:"encoding"`
}

//GitSearchAPIResponse : search response format
type GitSearchAPIResponse struct {
	TotalCount        int             `json:"total_count"`
	IncompleteResults bool            `json:"incomplete_results"`
	Items             []GitSearchItem `json:"items"`
}

func logErr(err error) {
	fmt.Println("[ERROR] " + err.Error())
	utils.ErrorLogger.Println(err.Error())
	return
}

func logInfo(info string) {
	utils.InfoLogger.Println(info)
	return
}

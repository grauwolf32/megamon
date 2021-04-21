package github

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/megamon/core/leaks/stage"
)

func buildGitSearchQuery(keyword string, lang string, infile bool) (query string) {
	query = keyword
	if infile {
		query += "+in:file"
	}

	if lang != "" {
		query += "+language:" + lang
	}

	return query
}

func buildGitSearchRequest(query string, offset int, token string) (req *http.Request, err error) {
	var requestBody bytes.Buffer
	url := fmt.Sprintf("https://api.github.com/search/code?q=%s&per_page=100&page=%d", query, offset)
	req, err = http.NewRequest("GET", url, &requestBody)

	if err != nil {
		return
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Accept-Encoding", "deflate, gzip;q=1.0, *;q=0.5")
	req.Header.Set("Connection", "close")
	return
}

//SearchStage : type of the stage interface
type SearchStage struct {
}

//BuildRequests : generate search requests
func (s *SearchStage) BuildRequests() (res chan stage.Request, err error) {
	return
}

//CheckResponse : check reponse
func (s *SearchStage) CheckResponse(resp stage.Response, reqCount int) (res int) {
	return
}

//ProcessResponse : process search response
func (s *SearchStage) ProcessResponse(resp []byte) (err error) {
	return
}

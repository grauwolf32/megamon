package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/megamon/core/config"
	"github.com/megamon/core/leaks/db"
	"github.com/megamon/core/leaks/stage"
)

func buildFetchRequest(url, token string) (*http.Request, error) {
	var requestBody bytes.Buffer
	req, err := http.NewRequest("GET", url, &requestBody)
	fmt.Println(url)

	if err != nil {
		return &http.Request{}, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Accept-Encoding", "deflate, gzip;q=1.0, *;q=0.5")
	return req, err
}

//FetchStage struct for the interface
type FetchStage struct {
	ReportHashes map[int][20]byte
	Manager      db.Manager
}

//Init : constructor
func (s *FetchStage) Init() {
	s.ReportHashes = make(map[int][20]byte)
}

//BuildRequests : generate search requests
func (s *FetchStage) BuildRequests() (reqQueue chan stage.Request, err error) {
	tokens := config.Settings.Github.Tokens
	reports, err := s.Manager.SelectReportByStatus("github", "processing")
	if err != nil {
		return
	}

	for id, report := range reports {
		s.ReportHashes[id] = report.ShaHash
		var gitSearchItem GitSearchItem
		err = json.Unmarshal(report.Data, &gitSearchItem)
		if err != nil {
			logErr(err)
			continue
		}

		token := tokens[id%len(tokens)]
		req, err := buildFetchRequest(gitSearchItem.GitURL, token)

		if err != nil {
			logErr(err)
			continue
		}
		reqQueue <- stage.Request{ID: id, Req: req}
	}
	return
}

//CheckResponse : check reponse
func (s *FetchStage) CheckResponse(resp stage.Response, reqCount int) (res int) {
	switch resp.Resp.StatusCode {
	case 200:
		return stage.OK
	case 403:
		fallthrough
	default:
		if reqCount < stage.MAXRETRIES {
			return stage.WAIT
		}
		return stage.SKIP
	}
}

//ProcessResponse : process search response
func (s *FetchStage) ProcessResponse(resp []byte, RequestID int) (err error) {
	var gitFetchItem GitFetchItem
	err = json.Unmarshal(resp, &gitFetchItem)
	if err != nil {
		logErr(err)
		return
	}

	var decoded []byte
	if gitFetchItem.Encoding == "base64" {
		/* Here is some magick: it seems that json automatically decode base64 encoding... */
		decoded = gitFetchItem.Content

	} else {
		err = fmt.Errorf("Fetcher.ProcessResponse: Unknown encoding: %s", gitFetchItem.Encoding)
		logErr(err)
		return
	}

	filePrefix := config.Settings.LeakGlobals.ContentDir
	filename := fmt.Sprintf("%s%x", filePrefix, s.ReportHashes[RequestID])

	err = ioutil.WriteFile(filename, decoded, 0644)
	if err != nil {
		logErr(err)
		return
	}

	return
}

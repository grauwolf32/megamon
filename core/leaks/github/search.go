package github

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/megamon/core/config"
	"github.com/megamon/core/leaks/db"
	"github.com/megamon/core/leaks/helpers"
	"github.com/megamon/core/leaks/stage"
	"github.com/megamon/core/utils"
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
	RequestParams map[int]gitRequestParams
	Manager       db.Manager
}

//Init : constructor
func (s *SearchStage) Init() {
	s.RequestParams = make(map[int]gitRequestParams)
	//TODO Init Manager
}

//BuildRequests : generate search requests
func (s *SearchStage) BuildRequests() (reqQueue chan stage.Request, err error) {
	defer close(reqQueue)
	keywords := config.Settings.LeakGlobals.Keywords
	langs := config.Settings.Github.Languages
	tokens := config.Settings.Github.Tokens
	var id int

	// nQueries := len(keywords) * len(langs)
	for i, lang := range langs {
		for j, keyword := range keywords {
			query := buildGitSearchQuery(keyword, lang, false)
			token := tokens[(i*len(keywords)+j)%len(tokens)]

			offset := 0
			req, err := buildGitSearchRequest(query, offset, token)

			if err != nil {
				utils.ErrorLogger.Println(err.Error())
				continue
			}

			// RateLimiter
			resp, err := utils.DoRequest(req)
			if err != nil {
				logErr(err)
				continue
			}

			bodyReader, err := utils.GetBodyReader(resp)
			if err != nil {
				logErr(err)
				continue
			}

			body, err := ioutil.ReadAll(bodyReader)
			bodyReader.Close()

			var githubResponse GitSearchAPIResponse
			err = json.Unmarshal(body, &githubResponse)
			if err != nil {
				logErr(err)
				continue
			}

			n := int(githubResponse.TotalCount / MAXRESPONSEITEMS)
			if githubResponse.TotalCount%MAXRESPONSEITEMS != 0 {
				n++
			}

			if n > MAXOFFSET {
				n = MAXOFFSET
			}

			for offset := 0; offset < n; offset++ {
				token := tokens[id%len(tokens)]
				req, err := buildGitSearchRequest(query, offset, token)

				if err != nil {
					utils.ErrorLogger.Println(err.Error())
					continue
				}

				reqQueue <- stage.Request{ID: id, Req: req}
				s.RequestParams[id] = gitRequestParams{query, keyword, offset}
				id++
			}
		}
	}

	return
}

//CheckResponse : check reponse
func (s *SearchStage) CheckResponse(resp stage.Response, reqCount int) (res int) {
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
func (s *SearchStage) ProcessResponse(resp []byte) (err error) {
	var githubResponse GitSearchAPIResponse
	err = json.Unmarshal(resp, &githubResponse)
	if err != nil {
		return
	}

	for _, gihubResponseItem := range githubResponse.Items {
		ShaHash, err := hex.DecodeString(gihubResponseItem.ShaHash)
		exist, err := s.Manager.CheckReportDuplicate(ShaHash)

		if err != nil {
			logErr(err)
			continue
		}

		if exist {
			continue
		}

		var report helpers.Report
		report.Type = "github"
		report.Status = "processing"
		report.Time = time.Now().Unix()
		copy(report.ShaHash[:], ShaHash[:20])

		if err != nil {
			return err
		}

		data, err := json.Marshal(gihubResponseItem)
		if err != nil {
			return err
		}

		report.Data = data
		_, err = s.Manager.InsertReport(report)

		if err != nil {
			return err
		}
	}

	return
}

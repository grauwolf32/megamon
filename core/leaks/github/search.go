package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/megamon/core/leaks/models"
	"github.com/megamon/core/leaks/stage"
	"github.com/megamon/core/utils"
	"golang.org/x/time/rate"
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
	logInfo(fmt.Sprintf("building search request: %s %d %s", query, offset, token[:4]))

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
	Manager       models.Manager
}

//Init : constructor
func (s *SearchStage) Init() (err error) {
	s.RequestParams = make(map[int]gitRequestParams)
	err = s.Manager.Init()
	return
}

//Close : destructor
func (s *SearchStage) Close() {
	s.Manager.Close()
	return
}

//GetDBManager : stage interface realization
func (s *SearchStage) GetDBManager() models.Manager {
	return s.Manager
}

//BuildRequests : generate search requests
func (s *SearchStage) BuildRequests(reqQueue chan stage.Request) (err error) {
	keywords, err := s.Manager.SelectKeywordByType(models.KWSEARCHABLE)
	if err != nil {
		logErr(err)
		return
	}

	tokens := utils.Settings.Github.Tokens
	desiredRate := rate.Limit(utils.Settings.Github.RequestRate) * rate.Every(time.Second)
	rl := rate.NewLimiter(desiredRate, 1)
	ctx := context.Background()
	id := 0

	// nQueries := len(keywords) * len(langs)
	for i, lang := range Langs {
		for j, keyword := range keywords {
			query := buildGitSearchQuery(keyword.Value, lang, false)
			token := tokens[(i*len(keywords)+j)%len(tokens)]

			offset := 0
			req, err := buildGitSearchRequest(query, offset, token)

			if err != nil {
				logErr(err)
				continue
			}

			_ = rl.Wait(ctx)
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
					logErr(err)
					continue
				}

				reqQueue <- stage.Request{ID: id, Req: req}
				s.RequestParams[id] = gitRequestParams{query, keyword.Value, offset}
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
func (s *SearchStage) ProcessResponse(resp []byte, requestID int) (err error) {
	logInfo(fmt.Sprintf("processing search API response from request : %d", requestID))

	var githubResponse GitSearchAPIResponse
	err = json.Unmarshal(resp, &githubResponse)
	if err != nil {
		return
	}

	for _, gihubResponseItem := range githubResponse.Items {
		exist, err := s.Manager.CheckReportDuplicate(gihubResponseItem.ShaHash)

		if err != nil {
			logErr(err)
			continue
		}

		if exist {
			continue
		}

		var report models.Report
		report.Type = "github"
		report.Status = stage.PROCESSED
		report.Time = time.Now().Unix()
		report.ShaHash = gihubResponseItem.ShaHash

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

package gist

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/megamon/core/leaks/github"
	"github.com/megamon/core/leaks/models"
	"github.com/megamon/core/leaks/stage"
	"github.com/megamon/core/utils"
	"golang.org/x/time/rate"
)

//Stage : Stage interface
type Stage struct {
	Manager       models.Manager
	RequestParams map[int]gistRequestParams
}

type gistRequestParams struct {
	keyword string
	page    int
}

//Init : constructor
func (s *Stage) Init() (err error) {
	s.RequestParams = make(map[int]gistRequestParams)
	err = s.Manager.Init()
	return
}

//Close : destructor
func (s *Stage) Close() {
	s.Manager.Close()
	return
}

//GetDBManager : stage interface realization
func (s *Stage) GetDBManager() models.Manager {
	return s.Manager
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

func gistSearchURL(query string, page int) string {
	URL := fmt.Sprintf("https://gist.github.com/search?p=%d&q=%s&ref=searchresults&s=updated", page, query)
	return URL
}

func buildSearchRequest(query string, page int, token string) (*http.Request, error) {
	url := gistSearchURL(query, page)
	logInfo(fmt.Sprintf("building gist search request: %s %s", url, token[:4]))

	var requestBody bytes.Buffer
	req, err := http.NewRequest("GET", url, &requestBody)

	if err != nil {
		return &http.Request{}, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Accept-Encoding", "deflate, gzip;q=1.0, *;q=0.5")
	return req, err
}

//BuildRequests : generate search requests
func (s *Stage) BuildRequests(reqQueue chan stage.Request) (err error) {
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
GENREQ:
	for i, keyword := range keywords {
		token := tokens[i%len(tokens)]
		nPages := []int{0, 5, 10, 20, 50, 100}
		nToLoad := 5

		for _, page := range nPages {
			req, err := buildSearchRequest(keyword.Value, page, token)

			if err != nil {
				logErr(err)
				continue GENREQ
			}

			_ = rl.Wait(ctx)
			resp, err := utils.DoRequest(req)
			if err != nil {
				logErr(err)
				continue GENREQ
			}

			bodyReader, err := utils.GetBodyReader(resp)
			if err != nil {
				logErr(err)
				continue GENREQ
			}

			body, err := ioutil.ReadAll(bodyReader)
			bodyReader.Close()

			if strings.Contains(string(body), "We couldn’t find any gists matching") {
				nToLoad = page
				break
			}
		}

		logInfo(fmt.Sprintf("loading %d pages for gist %s\n", nToLoad, keyword.Value))

		for offset := 0; offset < nToLoad; offset++ {
			token := tokens[id%len(tokens)]
			req, err := buildSearchRequest(keyword.Value, offset, token)
			if err != nil {
				logErr(err)
				continue
			}

			reqQueue <- stage.Request{ID: id, Req: req}
			s.RequestParams[id] = gistRequestParams{keyword: keyword.Value, page: offset}
			id++
		}
	}

	return
}

//CheckResponse : check reponse
func (s *Stage) CheckResponse(resp stage.Response, reqCount int) (res int) {
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
func (s *Stage) ProcessResponse(resp []byte, requestID int) (err error) {
	logInfo(fmt.Sprintf("processing search API response from request : %d", requestID))

	if strings.Contains(string(resp), "We couldn’t find any gists matching") {
		return
	}

	shaHash := sha1.Sum(resp)
	var report models.Report
	report.Data = resp
	report.Type = "gist"
	report.Time = time.Now().Unix()
	report.ShaHash = fmt.Sprintf("%x", shaHash)
	report.Status = stage.FETCHED

	_, err = s.Manager.InsertReport(report)

	if err != nil {
		return
	}

	filePrefix := utils.Settings.LeakGlobals.ContentDir
	filename := fmt.Sprintf("%s%s", filePrefix, report.ShaHash)

	err = ioutil.WriteFile(filename, resp, 0644)
	if err != nil {
		logErr(err)
		return
	}

	return
}

//GetTextsToProcess : produce report texts
func (s *Stage) GetTextsToProcess(textQueue chan stage.ReportText) (err error) {
	logInfo("generating texts for processing")
	reports, err := s.Manager.SelectReportByStatus("gist", stage.FETCHED)
	filePrefix := utils.Settings.LeakGlobals.ContentDir

	if err != nil {
		logErr(err)
		return
	}

	for _, report := range reports {
		filename := fmt.Sprintf("%s%s", filePrefix, report.ShaHash)
		logInfo(fmt.Sprintf("generating fragments for %s", filename))

		fileData, err := utils.ReadFile(filename)
		if err != nil {
			logErr(err)
			continue
		}

		textQueue <- stage.ReportText{ReportID: report.ID, Text: string(fileData)}
	}

	return
}

//ProcessTextFragment : stage interface realization
func (s *Stage) ProcessTextFragment(fragment models.TextFragment) (err error) {
	logInfo(fmt.Sprintf("processing fragment %s", fragment.ShaHash[:4]))
	exist, err := s.Manager.CheckTextFragmentDuplicate(fragment.ShaHash)
	if err != nil {
		return
	}
	if !exist {
		fragment.Type = "gist"
		_, err = s.Manager.InsertTextFragment(&fragment)
		return
	}
	return
}

//RunGistStage : main function
func RunGistStage(ctx context.Context) (err error) {
	var gistStage Stage
	gistStage.Init()
	var rl github.RateLimiter
	rl.Init()

	err = stage.RunStage(ctx, &gistStage, &rl, 1, 1, 2)
	gistStage.Close()

	if err != nil {
		logErr(err)
		return
	}

	err = github.UpdateState(stage.FRAGMENTED, stage.NEW, "gist")
	if err != nil {
		logErr(err)
	}

	return
}

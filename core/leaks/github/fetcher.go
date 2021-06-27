package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/megamon/core/leaks/models"
	"github.com/megamon/core/leaks/stage"
	"github.com/megamon/core/utils"
)

func buildFetchRequest(url, token string) (*http.Request, error) {
	logInfo(fmt.Sprintf("building fetch request: %s %s", url, token[:4]))

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

//FetchStage struct for the interface
type FetchStage struct {
	ReportHashes map[int]string
	ReportIDs    map[int]int
	Manager      models.Manager
}

//Init : constructor
func (s *FetchStage) Init() (err error) {
	s.ReportHashes = make(map[int]string)
	s.ReportIDs = make(map[int]int)
	err = s.Manager.Init()
	return
}

//Close : destructor
func (s *FetchStage) Close() {
	s.Manager.Close()
	return
}

//BuildRequests : generate search requests
func (s *FetchStage) BuildRequests(reqQueue chan stage.Request) (err error) {
	tokens := utils.Settings.Github.Tokens
	reports, err := s.Manager.SelectReportByStatus("github", stage.PROCESSED)
	if err != nil {
		return
	}

	for id, report := range reports {
		s.ReportHashes[id] = report.ShaHash
		s.ReportIDs[id] = report.ID

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

//GetDBManager : stage interface realization
func (s *FetchStage) GetDBManager() models.Manager {
	return s.Manager
}

//ProcessResponse : process search response
func (s *FetchStage) ProcessResponse(resp []byte, RequestID int) (err error) {
	logInfo(fmt.Sprintf("processing fetch response from request : %d", RequestID))

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

	filePrefix := utils.Settings.LeakGlobals.ContentDir
	filename := fmt.Sprintf("%s%s", filePrefix, s.ReportHashes[RequestID])

	err = ioutil.WriteFile(filename, decoded, 0644)
	if err != nil {
		logErr(err)
		return
	}

	reportID := s.ReportIDs[RequestID]
	s.Manager.UpdateReportStatus(reportID, stage.FETCHED)
	return
}

//GetTextsToProcess : produce report texts
func (s *FetchStage) GetTextsToProcess(textQueue chan stage.ReportText) (err error) {
	logInfo("generating texts for processing")
	reports, err := s.Manager.SelectReportByStatus("github", stage.FETCHED)
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
func (s *FetchStage) ProcessTextFragment(fragment models.TextFragment) (err error) {
	logInfo(fmt.Sprintf("processing fragment %s", fragment.ShaHash[:4]))
	exist, err := s.Manager.CheckTextFragmentDuplicate(fragment.ShaHash)
	if err != nil {
		return
	}
	if !exist {
		fragment.Type = "github"
		_, err = s.Manager.InsertTextFragment(&fragment)
		return
	}
	return
}

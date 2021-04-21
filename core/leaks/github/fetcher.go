package github

import (
	"bytes"
	"fmt"
	"net/http"

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
}

//BuildRequests : generate search requests
func (s *FetchStage) BuildRequests() (res chan stage.Request, err error) {
	return
}

//CheckResponse : check reponse
func (s *FetchStage) CheckResponse(resp stage.Response, reqCount int) (res int) {
	return
}

//ProcessResponse : process search response
func (s *FetchStage) ProcessResponse(resp []byte) (err error) {
	return
}

package github

import (
	"time"

	"golang.org/x/time/rate"
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

//RateLimiter : rate limit request to github
type RateLimiter struct {
	RequestRate float64
	Duration    time.Duration
	Limiter     *rate.Limiter
}

// Langs : supported langs for search
var Langs = [...]string{"", "C", "C#", "C++", "CoffeeScript", "CSS", "Dart", "DM", "Elixir", "Go", "Groovy", "HTML", "Java",
	"JavaScript", "Kotlin", "Objective-C", "Perl", "PHP", "PowerShell", "Python", "Ruby", "Rust",
	"Scala", "Shell", "Swift", "TypeScript", "CSV", "JSON", "Makefile", "Markdown", "YAML", "XML",
	"Diff", "Erlang", "GraphQL", "Jupyter+Notebook", "Lua", "Protocol+Buffer", "Public+Key", "SQL",
	"SSH+Config", "Text"}

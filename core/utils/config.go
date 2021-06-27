package utils

import "regexp"

//RejectRule : description of reject rule
type RejectRule struct {
	ID   int    `json:"id"`
	Rule string `json:"rule"`
	Name string `json:"name"`
	Expr *regexp.Regexp
}

//Keyword : auxilary data type
type Keyword struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
	Type  int    `json:"type"`
}

//GlobalSettings : settings for the whole project
type GlobalSettings struct {
	Github           githubSettings        `yaml:"github" json:"github"`
	DBCredentials    DBCredentialsSettings `yaml:"db_redentials" json:"db_redentials"`
	LeakGlobals      leakGlobalsSettings   `yaml:"globals" json:"globals"`
	AdminCredentials webAdminSettings      `yaml:"admin_credentials" json:"admin_credentials"`
}

//Blacklist : black list
type Blacklist struct {
	Blacklist []string `yaml:"blacklist" json:"blacklist"`
}

//DBCredentialsSettings : database credentials
type DBCredentialsSettings struct {
	Database   string `yaml:"database" json:"database"`
	Name       string `yaml:"name" json:"name"`
	Password   string `yaml:"password" json:"password"`
	DBHostName string `yaml:"db_hostname"`
}

type githubSettings struct {
	Tokens      []string  `yaml:"tokens" json:"tokens"`
	Langs       Blacklist `yaml:"langs" json:"langs"`
	RequestRate float64   `yaml:"request_rate" json:"request_rate"`
}

type leakGlobalsSettings struct {
	Version  float32
	Keywords map[string]Keyword    `json:"keywords"`
	Rules    map[string]RejectRule `json:"rules"`

	ContentDir string `yaml:"content_dir" json:"content_dir"`
	LogDir     string `yaml:"log_dir" json:"log_dir"`
	LogFile    string `yaml:"log_file" json:"log_file"`
}

type webAdminSettings struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

//Settings : global instance of settings for the project
var Settings GlobalSettings

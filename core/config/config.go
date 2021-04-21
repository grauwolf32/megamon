package config

//GlobalSettings : settings for the whole project
type GlobalSettings struct {
	Github           githubSettings        `json:"github"`
	DBCredentials    DBCredentialsSettings `json:"db_redentials"`
	LeakGlobals      leakGlobalsSettings   `json:"globals"`
	AdminCredentials webAdminSettings      `json:"admin_credentials"`
}

//DBCredentialsSettings : database credentials
type DBCredentialsSettings struct {
	Database string `json:"database"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type githubSettings struct {
	Tokens    []string `json:"tokens"`
	Languages []string `json:"langs"`
}

type leakGlobalsSettings struct {
	Keywords    []string `json:"keywords"`
	ExcludeList []string `json:"exclude"`
	ContentDir  string   `json:"content_dir"`
}

type webAdminSettings struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//Settings : global instance of settings for the project
var Settings GlobalSettings

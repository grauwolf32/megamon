package utils

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
	Database string `yaml:"database" json:"database"`
	Name     string `yaml:"name" json:"name"`
	Password string `yaml:"password" json:"password"`
}

type githubSettings struct {
	Tokens      []string  `yaml:"tokens" json:"tokens"`
	Langs       Blacklist `yaml:"langs" json:"langs"`
	RequestRate float64   `yaml:"request_rate" json:"request_rate"`
}

type leakGlobalsSettings struct {
	Version    float32
	Keywords   []string `yaml:"keywords" json:"keywords"`
	ContentDir string   `yaml:"content_dir" json:"content_dir"`
	LogDir     string   `yaml:"log_dir" json:"log_dir"`
	LogFile    string   `yaml:"log_file" json:"log_file"`
}

type webAdminSettings struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

//Settings : global instance of settings for the project
var Settings GlobalSettings

package utils

//GlobalSettings : settings for the whole project
type GlobalSettings struct {
	Github           githubSettings        `yaml:"github"`
	DBCredentials    DBCredentialsSettings `yaml:"db_redentials"`
	LeakGlobals      leakGlobalsSettings   `yaml:"globals"`
	AdminCredentials webAdminSettings      `yaml:"admin_credentials"`
}

//Blacklist : black list
type Blacklist struct {
	Blacklist []string `yaml:"blacklist"`
}

//DBCredentialsSettings : database credentials
type DBCredentialsSettings struct {
	Database string `yaml:"database"`
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}

type githubSettings struct {
	Tokens      []string  `yaml:"tokens"`
	Langs       Blacklist `yaml:"langs"`
	RequestRate float64   `yaml:"request_rate"`
}

type leakGlobalsSettings struct {
	Version    float32
	Keywords   []string `yaml:"keywords"`
	ContentDir string   `yaml:"content_dir"`
	LogDir     string   `yaml:"log_dir"`
	LogFile    string   `yaml:"log_file"`
}

type webAdminSettings struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

//Settings : global instance of settings for the project
var Settings GlobalSettings

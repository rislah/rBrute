package config

type Config struct {
	Settings Settings `yaml:"settings"`
	Stages   Stages   `yaml:"stages"`
}

type Stages struct {
	PreLogin      []PreLoginStage `yaml:"preLogin"`
	Login         LoginStage      `yaml:"login"`
	GlobalHeaders []Header        `yaml:"globalHeaders"`
}

type Settings struct {
	BotCount          int    `yaml:"botCount"`
	UnbanProxiesAfter int    `yaml:"unbanProxiesAfter"`
	ProxyMaxRetries   int    `yaml:"proxyMaxRetries"`
	ConfigName        string `yaml:"configName"`
	UseProxy          bool   `yaml:"useProxy"`
}

type Stage struct {
	URL     string   `yaml:"url"`
	Method  Method   `yaml:"method"`
	Body    string   `yaml:"body"`
	Headers []Header `yaml:"headers"`
}

type LoginStage struct {
	Stage    `yaml:",inline"`
	Keywords Keywords `yaml:"keywords"`
}

type Keywords struct {
	Success struct {
		Text []string `yaml:"text"`
	} `yaml:"success"`
	Failure struct {
		Text []string `yaml:"text"`
	} `yaml:"failure"`
}

type PreLoginStage struct {
	Stage           `yaml:",inline"`
	VariablesToSave []VariablesToSave `yaml:"variablesToSave"`
}

type Method string

func (m Method) ToString() string {
	return string(m)
}

const (
	GET  Method = "GET"
	POST Method = "POST"
)

type Header struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type VariablesToSave struct {
	Name           string `yaml:"name"`
	LeftDelimiter  string `yaml:"leftDelimiter"`
	RightDelimiter string `yaml:"rightDelimiter"`
}

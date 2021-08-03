package logger

import (
	"github.com/rislah/rBrute/channels"
	"net/http"
)

type LoggerContext struct {
	preLoginRequests []*http.Request
	loginRequest     *http.Request
	initVariables    []map[string]string
	foundVariables   map[string]string
	keywords         []string
	responseBody     string
	credentials      *channels.Credentials
}

func NewLoggerContext() *LoggerContext {
	return &LoggerContext{
		foundVariables: make(map[string]string),
	}
}

func (l *LoggerContext) GetCredentials() *channels.Credentials {
	return l.credentials
}

func (l *LoggerContext) AddCredentials(creds *channels.Credentials) {
	l.credentials = creds
}

func (l *LoggerContext) GetPreLoginRequests() []*http.Request {
	return l.preLoginRequests
}

func (l *LoggerContext) AddPreLoginRequest(request *http.Request) {
	l.preLoginRequests = append(l.preLoginRequests, request)
}

func (l *LoggerContext) GetLoginRequest() *http.Request {
	return l.loginRequest
}

func (l *LoggerContext) AddLoginRequest(request *http.Request) {
	l.loginRequest = request
}

func (l *LoggerContext) GetKeywords() []string {
	return l.keywords
}

func (l *LoggerContext) AddKeyword(keyword string) {
	l.keywords = append(l.keywords, keyword)
}

func (l *LoggerContext) GetResponseBody() string {
	return l.responseBody
}

func (l *LoggerContext) AddResponseBody(response string) {
	l.responseBody = response
}

func (l *LoggerContext) GetInitVariables() []map[string]string {
	return l.initVariables
}

func (l *LoggerContext) AddInitVariables(variables map[string]string) {
	initVariables := []map[string]string{}
	for k, v := range variables {
		entry := map[string]string{}
		entry[k] = v
		initVariables = append(initVariables, entry)
	}
	l.initVariables = initVariables
}

func (l *LoggerContext) GetFoundVariables() map[string]string {
	return l.foundVariables
}

func (l *LoggerContext) AddFoundVariables(name, value string) {
	l.foundVariables[name] = value
}

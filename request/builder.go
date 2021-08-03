package request

import (
	"bytes"
	"log"
	"net/http"

	"github.com/rislah/rBrute/config"
	"github.com/rislah/rBrute/logger"
)

type builder struct {
	loggerContext *logger.LoggerContext
	stages        *config.Stages
	variables     *Variables
}

func NewBuilder(lx *logger.LoggerContext, stages *config.Stages, v *Variables) *builder {
	return &builder{
		loggerContext: lx,
		stages:        stages,
		variables:     v,
	}
}

func (b *builder) BuildPreLoginRequests() []map[*config.PreLoginStage]*http.Request {
	if b.stages.PreLogin == nil || len(b.stages.PreLogin) == 0 {
		return nil
	}

	requests := []map[*config.PreLoginStage]*http.Request{}
	for _, preLoginStage := range b.stages.PreLogin {
		request, err := b.buildRequest(preLoginStage.Method, preLoginStage.Headers, preLoginStage.URL, preLoginStage.Body)
		if err != nil {
			log.Fatal(err)
		}

		b.loggerContext.AddPreLoginRequest(request)

		elem := make(map[*config.PreLoginStage]*http.Request)
		elem[&preLoginStage] = request

		requests = append(requests, elem)
	}
	return requests
}

func (b *builder) BuildLoginRequest() *http.Request {
	request, err := b.buildRequest(b.stages.Login.Method, b.stages.Login.Headers, b.stages.Login.URL, b.stages.Login.Body)
	b.loggerContext.AddLoginRequest(request)
	if err != nil {
		log.Fatal(err)
	}
	return request
}

func (b *builder) buildRequest(method config.Method, headers []config.Header, url, body string) (*http.Request, error) {
	request, err := http.NewRequest(method.ToString(), b.variables.Replace(url), bytes.NewBuffer([]byte(b.variables.Replace(body))))
	if err != nil {
		return nil, err
	}
	b.addHeaders(request, b.stages.GlobalHeaders)
	b.addHeaders(request, headers)
	return request, nil
}

func (b *builder) addHeaders(req *http.Request, headers []config.Header) {
	for _, v := range headers {
		req.Header.Set(v.Key, v.Value)
	}
}

package rbrute

import (
	"context"
	"github.com/rislah/rBrute/channels"
	"sync"

	"github.com/rislah/rBrute/config"
	"github.com/rislah/rBrute/logger"
	"github.com/rislah/rBrute/request"
	"github.com/rs/xid"
)

type Worker struct {
	ctx         context.Context
	status      int
	logger      *logger.Logger
	credsStream <-chan *channels.Credentials
	proxyStream <-chan *channels.Proxy
	proxyGen    *channels.ProxyFO
	stages      config.Stages
	useProxy    bool
	maxRetryCount int
	mutex       sync.RWMutex
}

func NewWorker(
	ctx context.Context,
	credsStream <-chan *channels.Credentials,
	proxyStream <-chan *channels.Proxy,
	log *logger.Logger,
	stages config.Stages,
	useProxy bool,
	proxyGen *channels.ProxyFO,
	maxRetryCount int,
) *Worker {
	return &Worker{
		ctx:           ctx,
		credsStream:   credsStream,
		proxyStream:   proxyStream,
		logger:        log,
		stages:        stages,
		maxRetryCount: maxRetryCount,
		useProxy:      useProxy,
		proxyGen:      proxyGen,
	}
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	client := request.BuildClient()
	for {
		select {
		case <-w.ctx.Done():
			return
		case creds, ok := <-w.credsStream:
			if !ok {
				return
			}

			loggerContext := logger.NewLoggerContext()
			loggerContext.AddCredentials(creds)

			name := xid.New().String()
			vars := request.NewVariables(loggerContext)
			rb := request.NewBuilder(loggerContext, &w.stages, &vars)
			rr := request.NewRunner(name, w.logger, loggerContext, &vars, w.proxyStream, creds, w.useProxy, client, w.maxRetryCount)

			preloginRequests := rb.BuildPreLoginRequests()
			if !rr.ProcessPreLoginRequests(preloginRequests) {
				w.logger.PrintFailedMessage(creds)
				continue
			}

			keywordsChecker := request.NewKeywords(loggerContext, &w.stages.Login.Keywords)
			loginRequest := rb.BuildLoginRequest()
			msg, ok := rr.ProcessLoginRequest(loginRequest, &keywordsChecker)
			if !ok {
				w.logger.PrintFailedMessage(creds)
				continue
			}
			w.logger.PrintSuccessMessage(creds, msg)

			if w.useProxy {
				inUseProxy := rr.GetInUseProxy()
				w.proxyGen.UpdateWorkingProxy(inUseProxy)
			}
			<-w.logger.LogContextToFile(w.ctx, loggerContext)

		}
	}
}

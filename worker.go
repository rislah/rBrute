package rbrute

import (
	"context"
	"sync"

	"github.com/rislah/rBrute/checker"
	"github.com/rislah/rBrute/combolist"
	"github.com/rislah/rBrute/config"
	"github.com/rislah/rBrute/logger"
	"github.com/rislah/rBrute/proxy"
	"github.com/rislah/rBrute/request"
	"github.com/rislah/rBrute/variables"
)

type Worker struct {
	ctx         context.Context
	logger      *logger.Logger
	credsStream <-chan *combolist.Credentials
	proxyStream <-chan *proxy.Proxy
	proxyGen    *proxy.Generator
	stages      config.Stages
	useProxy    bool
	mutex       sync.RWMutex
}

func NewWorker(
	ctx context.Context,
	credsStream <-chan *combolist.Credentials,
	proxyStream <-chan *proxy.Proxy,
	log *logger.Logger,
	stages config.Stages,
	useProxy bool,
	proxyGen *proxy.Generator,
) *Worker {
	return &Worker{
		ctx:         ctx,
		credsStream: credsStream,
		proxyStream: proxyStream,
		proxyGen:    proxyGen,
		logger:      log,
		stages:      stages,
		useProxy:    useProxy,
	}
}

func (w *Worker) start(wg *sync.WaitGroup) {
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

			vars := variables.NewVariables(loggerContext)
			rb := request.NewBuilder(loggerContext, &w.stages, &vars)
			rr := request.NewRunner(loggerContext, &vars, w.proxyStream, creds, w.useProxy, client)

			preloginRequests := rb.BuildPreLoginRequests()
			if !rr.ProcessPreLoginRequests(preloginRequests) {
				w.logger.PrintFailedMessage(creds)
				continue
			}

			keywordsChecker := checker.NewKeywords(loggerContext, &w.stages.Login.Keywords)
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

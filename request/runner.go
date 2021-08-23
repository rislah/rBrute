package request

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"github.com/rislah/rBrute/channels"

	"github.com/rislah/rBrute/config"
	"github.com/rislah/rBrute/logger"
	"golang.org/x/net/publicsuffix"
	"h12.io/socks"
)

type Runner struct {
	loggerContext    *logger.LoggerContext
	logger           *logger.Logger
	proxyStream      <-chan *channels.Proxy
	inUseCredentials *channels.Credentials
	variables        *Variables
	useProxy         bool
	retrier          retrier
	name             string
	inUseProxy       *channels.Proxy
	inUseClient      *http.Client
	mutex            sync.RWMutex
}

func NewRunner(name string, l *logger.Logger, lx *logger.LoggerContext, v *Variables, ps <-chan *channels.Proxy, iuc *channels.Credentials, up bool,
	client *http.Client, maxRetryCount int) *Runner {
	runner := &Runner{
		name:             name,
		logger:           l,
		loggerContext:    lx,
		retrier:          newRetrier(maxRetryCount),
		proxyStream:      ps,
		variables:        v,
		inUseCredentials: iuc,
		useProxy:         up,
		inUseClient:      client,
	}
	runner.setupClient()
	return runner
}

func BuildClient() *http.Client {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}

	jar, err := cookiejar.New(&options)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: time.Duration(3) * time.Second,
	}
	return client
}

func (r *Runner) setupClient() {
	if r.useProxy {
		r.setInUseProxy(<-r.proxyStream)
		r.setClientProxy()
	} else {
		r.clearProxyFromClient()
	}
}

func (r *Runner) clearProxyFromClient() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.inUseClient.Transport = nil
}

func (r *Runner) setInUseProxy(proxy *channels.Proxy) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.inUseProxy = proxy
}

func (r *Runner) setClientProxy() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	dial := socks.DialSocksProxy(socks.SOCKS4, r.inUseProxy.GetAddr())
	tr := &http.Transport{Dial: dial}
	r.inUseClient.Transport = tr
}

func (r *Runner) ProcessPreLoginRequests(stageRequests []map[*config.PreLoginStage]*http.Request) bool {
	if stageRequests == nil {
		return true
	}

	for _, v := range stageRequests {
		for cfg, req := range v {
			res := r.sendRequest(req)
			if res == nil {
				return false
			}

			strRes, err := responseToString(res)
			if err != nil {
				log.Fatal(err)
			}

			if !r.variables.FindAndSave(strRes, cfg.VariablesToSave) {
				r.logger.PrintStatusChange(r.name, r.inUseCredentials, r.inUseProxy, logger.FAILED)
				return false
			}
		}
	}
	return true
}

func (r *Runner) GetInUseProxy() *channels.Proxy {
	return r.inUseProxy
}

func (r *Runner) ProcessLoginRequest(request *http.Request, kc *Keywords) (string, bool) {
	response := r.sendRequest(request)
	resStr, err := responseToString(response)
	if err != nil {
		log.Fatal(err)
	}
	r.loggerContext.AddResponseBody(resStr)
	return kc.Check(resStr)
}

func (r *Runner) sendRequest(request *http.Request) *http.Response {
	var res *http.Response
	err := r.retrier.retry(func(attempt int) error {
		var err error
		res, err = r.inUseClient.Do(request)
		if attempt >= 1 {
			r.logger.PrintStatusChange(r.name, r.inUseCredentials, r.inUseProxy, logger.RETRYING, fmt.Sprintf("TRY %d/%d, SUCCESS COUNT: %d", attempt+1, r.retrier.maxRetryCount, r.inUseProxy.GetSuccessCount()))
		}
		r.inUseProxy.DecrementTotalSuccessCount()
		return err
	})
	if err != nil {
		if r.useProxy {
			r.inUseProxy.ChangeStatusToBanned()
			r.setInUseProxy(<-r.proxyStream)
			r.setClientProxy()
			return r.sendRequest(request)
		}
	}
	return res
}

func responseToString(response *http.Response) (string, error) {
	defer response.Body.Close()
	read, err := ioutil.ReadAll(response.Body)
	if err != nil && err != io.EOF {
		return "", err
	}
	return string(read), nil
}

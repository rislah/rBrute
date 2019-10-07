package request

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"github.com/rislah/rBrute/checker"
	"github.com/rislah/rBrute/combolist"
	"github.com/rislah/rBrute/config"
	"github.com/rislah/rBrute/logger"
	"github.com/rislah/rBrute/proxy"
	"github.com/rislah/rBrute/variables"
	"golang.org/x/net/publicsuffix"
	"h12.io/socks"
)

type Runner struct {
	lx               *logger.LoggerContext
	proxyStream      <-chan *proxy.Proxy
	variables        *variables.Variables
	useProxy         bool
	inUseCredentials *combolist.Credentials
	retrier          retrier

	mutex       sync.RWMutex
	inUseProxy  *proxy.Proxy
	inUseClient *http.Client
}

func NewRunner(lx *logger.LoggerContext, v *variables.Variables, ps <-chan *proxy.Proxy, iuc *combolist.Credentials, up bool, client *http.Client) *Runner {
	runner := &Runner{
		lx:               lx,
		retrier:          newRetrier(3),
		proxyStream:      ps,
		variables:        v,
		inUseCredentials: iuc,
		useProxy:         up,
		inUseClient:      client,
	}
	runner.setupClient(<-ps)
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

func (r *Runner) setupClient(proxy *proxy.Proxy) {
	if r.useProxy && proxy != nil {
		r.setInUseProxy(proxy)
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

func (r *Runner) setInUseProxy(proxy *proxy.Proxy) {
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

func (r *Runner) ProcessPreLoginRequests(requests []map[*config.PreLoginStage]*http.Request) bool {
	if requests == nil {
		return true
	}

	for _, elem := range requests {
		for cfg, req := range elem {
			res := r.sendRequest(req)
			if res == nil {
				return false
			}

			strRes, err := responseToString(res)
			if err != nil {
				log.Fatal(err)
			}

			if !r.variables.FindAndSave(strRes, cfg.VariablesToSave) {
				return false
			}
		}
	}
	return true
}

func (r *Runner) GetInUseProxy() *proxy.Proxy {
	return r.inUseProxy
}

func (r *Runner) ProcessLoginRequest(request *http.Request, kc *checker.Keywords) (string, bool) {
	response := r.sendRequest(request)
	resStr, err := responseToString(response)
	if err != nil {
		log.Fatal(err)
	}
	r.lx.AddResponseBody(resStr)
	return kc.Check(resStr)
}

func (r *Runner) sendRequest(request *http.Request) *http.Response {
	var res *http.Response
	err := r.retrier.retry(func(attempt int) error {
		var err error
		res, err = r.inUseClient.Do(request)
		return err
	})
	if err != nil {
		if r.useProxy {
			r.inUseProxy.ChangeStatusToBanned()
			r.setInUseProxy(<-r.proxyStream)
			r.setClientProxy()
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

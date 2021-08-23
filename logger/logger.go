package logger

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/rislah/rBrute/channels"
)

type Logger struct {
	logger                  *log.Logger
	resultsDir              string
	currentConfigResultsDir string
}

type Status int

const (
	RETRYING Status = iota
	SUCCESS
	FAILED
)

func (s Status) ToString() string {
	switch s {
	case SUCCESS:
		return "SUCCESS"
	case FAILED:
		return "FAILED"
	case RETRYING:
		return "RETRYING"
	default:
		return "PROCESSING"
	}
}

func NewLogger(resultsDir string) Logger {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	return Logger{
		logger:     logger,
		resultsDir: resultsDir,
	}
}

func (l *Logger) PrintSuccessMessage(credentials *channels.Credentials, keywordsCheckResult string) {
	l.logger.Println(fmt.Sprintf(
		"[%s] %s:%s - %s",
		aurora.BrightGreen("SUCCESS"),
		credentials.Username,
		credentials.Password,
		keywordsCheckResult,
	))
}

func (l *Logger) PrintStatusChange(name string, credentials *channels.Credentials, proxy *channels.Proxy, s Status, misc ...string) {
	l.logger.Println(fmt.Sprintf(
		"[%s] [BOT %s] [PROXY %s] [CREDENTIALS %s:%s] - %s %s",
		aurora.BrightBlack("STATUS CHANGE"),
		name,
		proxy.GetAddr(),
		credentials.Username,
		credentials.Password,
		s.ToString(),
		misc,
	))
}

func (l *Logger) PrintFailedMessage(credentials *channels.Credentials) {
	l.logger.Println(fmt.Sprintf(
		"[%s] %s:%s",
		aurora.BrightRed("FAILED"),
		credentials.Username,
		credentials.Password,
	))
}

func (l *Logger) Init(configName string) {
	now := time.Now()
	cdate := l.currentDate(now)
	fpath := filepath.Join(l.resultsDir, cdate, configName)
	l.createDirs(fpath)
	l.currentConfigResultsDir = fpath
}

func (l *Logger) LogContextToFile(ctx context.Context, loggerContext *LoggerContext) <-chan bool {
	doneStream := make(chan bool)
	go func() {
		defer close(doneStream)
		creds := loggerContext.GetCredentials()
		configPath := filepath.Join(l.currentConfigResultsDir, creds.Username)
		l.createDirs(configPath)

		logFilePath := filepath.Join(configPath, "log.txt")
		logFile := l.createFile(logFilePath)

		log.SetOutput(logFile)
		log.SetFlags(0)
		log.Print("==============================START===============================")
		log.Print(header("INIT VARIABLES"))
		for _, v := range loggerContext.GetInitVariables() {
			for k, v := range v {
				log.Println(k, v)
			}
		}
		log.Println(header("PRELOGIN REQUESTS"))
		for i, req := range loggerContext.GetPreLoginRequests() {
			log.Println(
				fmt.Sprintf("%d| URL: %s", i, req.URL.String()),
			)
			log.Println(
				fmt.Sprintf("%d| Method: %s", i, req.Method),
			)
			log.Println(
				fmt.Sprintf("%d| Header: %+v", i, req.Header),
			)
			log.Println(
				fmt.Sprintf("%d| Body: %+v", i, req.Body),
			)
			log.Println()
		}
		log.Println(header("FOUND VARIABLES"))
		for k, v := range loggerContext.GetFoundVariables() {
			log.Println(k, v)
		}
		log.Println(header("LOGIN REQUEST"))
		loginRequest := loggerContext.GetLoginRequest()
		log.Println(
			fmt.Sprintf("URL: %s", loginRequest.URL.String()),
		)
		log.Println()
		log.Println(
			fmt.Sprintf("Method: %s", loginRequest.Method),
		)
		log.Println()
		log.Println(
			fmt.Sprintf("Header: %+v", loginRequest.Header),
		)
		log.Println()
		resStr, err := ioutil.ReadAll(loginRequest.Body)
		if err == io.EOF {
		}

		log.Println(
			fmt.Sprintf("Body: %s", resStr),
		)
		log.Println(header("KEYWORDS"))
		for _, k := range loggerContext.GetKeywords() {
			log.Println(k)
		}
		logFile.Close()

		responseFilePath := filepath.Join(configPath, "response.html")
		hFile := l.createFile(responseFilePath)
		hFile.Truncate(0)
		hFile.Seek(0, 0)
		hFile.WriteString(loggerContext.GetResponseBody())
		hFile.Close()
	}()
	return doneStream
}

func header(name string) string {
	return fmt.Sprintf(`
=====================
 %s         
=====================
    `, name)
}

func (l *Logger) createFile(filePath string) *os.File {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func (l *Logger) createDirs(fpath string) {
	if _, serr := os.Stat(fpath); serr != nil {
		merr := os.MkdirAll(fpath, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
	}
}

func (l *Logger) currentDate(now time.Time) string {
	return fmt.Sprintf(
		"%02d-%02d-%d",
		now.Day(), now.Month(), now.Year(),
	)
}

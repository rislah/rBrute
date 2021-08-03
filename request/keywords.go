package request

import (
	"fmt"
	"strings"

	"github.com/rislah/rBrute/config"
	"github.com/rislah/rBrute/logger"
)

type Keywords struct {
	lx       *logger.LoggerContext
	keywords *config.Keywords
}

func NewKeywords(lx *logger.LoggerContext, keywords *config.Keywords) Keywords {
	return Keywords{
		lx:       lx,
		keywords: keywords,
	}
}

func (k Keywords) Check(body string) (string, bool) {
	str, failureFound := k.findKeywords(body, k.keywords.Failure.Text, func(s string) string {
		return fmt.Sprintf("--found failure keyword: %s--", s)
	})
	if failureFound {
		return str, false
	}

	str, successFound := k.findKeywords(body, k.keywords.Success.Text, func(s string) string {
		return fmt.Sprintf("--found success keyword: %s--", s)
	})
	if successFound {
		return str, true
	}
	return "", false
}

type printFn func(string) string

func (k Keywords) findKeywords(body string, keywords []string, fn printFn) (string, bool) {
	for _, keyword := range keywords {
		if strings.Contains(body, keyword) {
			str := fn(keyword)
			k.lx.AddKeyword(str)
			return str, true
		}
	}
	return "", false
}

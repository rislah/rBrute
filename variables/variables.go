package variables

import (
	"strings"

	"github.com/rislah/rBrute/config"
	"github.com/rislah/rBrute/logger"
)

type Variables struct {
	storage       *Storage
	loggerContext *logger.LoggerContext
}

func NewVariables(loggerContext *logger.LoggerContext) Variables {
	v := Variables{
		storage:       NewStorage(),
		loggerContext: loggerContext,
	}
	v.InitDefaultStorageVariables()
	v.loggerContext.AddInitVariables(v.storage.GetVariables())
	return v
}

func (vr Variables) GetStorage() *Storage {
	return vr.storage
}

func (vr Variables) InitDefaultStorageVariables() {
	vr.storage.InitDefaultVariables(vr.loggerContext.GetCredentials())
}

func (vr Variables) FindAndSave(line string, variablesToSave []config.VariablesToSave) bool {
	for _, vs := range variablesToSave {
		res := vr.getBetween(line, vs.LeftDelimiter, vs.RightDelimiter)
		if res == "" {
			return false
		}
		vr.loggerContext.AddFoundVariables(vs.Name, res)
		vr.storage.AddVariable(vs.Name, res)
	}
	return true
}

func (vr Variables) getBetween(str, left, right string) (res string) {
	leftIndex := strings.Index(str, left)
	if leftIndex == -1 {
		return
	}
	leftIndex += len(left)

	rightIndex := strings.Index(str, right)
	if rightIndex == -1 {
		return
	}

	res = str[leftIndex : leftIndex+rightIndex]
	return
}

func (vr Variables) Replace(line string) string {
	storage := vr.storage.GetVariables()
	for k, v := range storage {
		if strings.Contains(line, k) {
			line = vr.replace(line, k, v)
		}
	}
	return line
}

func (vr Variables) replace(str, substring, newstr string) string {
	return strings.Replace(str, substring, newstr, -1)
}

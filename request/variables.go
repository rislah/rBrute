package request

import (
	"strings"
	"sync"

	"github.com/rislah/rBrute/channels"

	"github.com/rislah/rBrute/config"
	"github.com/rislah/rBrute/logger"
)

type Variables struct {
	storage       *Storage
	loggerContext *logger.LoggerContext
}

type Storage struct {
	mutex     sync.RWMutex
	variables map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		variables: make(map[string]string),
	}
}

func (s *Storage) InitDefaultVariables(creds *channels.Credentials) {
	s.AddVariable("<username>", creds.Username)
	s.AddVariable("<password>", creds.Password)
}

func (s *Storage) AddVariable(key, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.variables[key] = value
}

func (s *Storage) GetVariables() map[string]string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.variables
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

package variables

import (
	"sync"

	"github.com/rislah/rBrute/combolist"
)

type Storage struct {
	mutex     sync.RWMutex
	variables map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		variables: make(map[string]string),
	}
}

func (s *Storage) InitDefaultVariables(creds *combolist.Credentials) {
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

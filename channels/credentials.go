package channels

import (
	"context"
	"log"
)

type Credentials struct {
	Username string
	Password string
}

type CredentialsFO struct {
	channel  chan *Credentials
	filePath string
}

func NewCredentialsFO(ch chan *Credentials, filePath string) CredentialsFO {
	return CredentialsFO{ch, filePath}
}

func (clo CredentialsFO) Produce(ctx context.Context) {
	lines, err := readLines(clo.filePath)
	if err != nil {
		log.Fatal(err)
	}
	var credentials []*Credentials
	for _, line := range lines {
		split := splitByColon(line)
		credentials = append(credentials, &Credentials{split[0], split[1]})
	}
	for _, c := range credentials {
		select {
		case <-ctx.Done():
			return
		case clo.channel <- c:
		}
	}
	defer close(clo.channel)
}

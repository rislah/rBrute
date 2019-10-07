package combolist

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type Credentials struct {
	Username string
	Password string
}

func newCredentials(username, password string) *Credentials {
	return &Credentials{username, password}
}

func NewCredentialsList(path string) []*Credentials {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer file.Close()
	br := bufio.NewReader(file)
	credentials := []*Credentials{}
	for {
		bytesLine, err := br.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		if bytesLine == nil {
			break
		}

		line := strings.TrimSpace(string(bytesLine))
		split := splitByColon(line)
		if split != nil {
			credentials = append(credentials, split)
		}
	}
	return credentials
}

func splitByColon(line string) *Credentials {
	split := strings.Split(line, ":")
	if len(split) != 2 {
		return nil
	}
	return newCredentials(split[0], split[1])
}

package channels

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
)

type ChannelInterface interface {
	Produce(ctx context.Context)
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	strs := []string{}
	buffer := bufio.NewReader(file)
	for {
		rd, err := buffer.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				if len(rd) != 0 {
					//
				} else {
					break
				}
			} else {
				panic(err)
			}
		}
		tr := strings.TrimSpace(string(rd))
		strs = append(strs, tr)
	}
	return strs, nil
}

func splitByColon(line string) []string {
	split := strings.Split(line, ":")
	if len(split) != 2 {
		return nil
	}
	return []string{split[0], split[1]}
}

package utils

import (
	"errors"
	"strings"
)

func ParseServicePath(path string) (string, string, error) {
	index := strings.LastIndex(path, "/")
	if index == 0 || index == -1 || !strings.HasPrefix(path, "/") {
		return "", "", errors.New("invalid path")
	}
	return path[1:index], path[index+1:], nil
}

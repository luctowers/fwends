package util

import (
	"errors"
	"os"
	"regexp"
	"strconv"
)

var r = regexp.MustCompile(`^fwends-backend-([0-9]+)$`)

func PodIndex() (int64, error) {
	host, err := os.Hostname()
	if err != nil {
		return 0, err
	}
	matches := r.FindStringSubmatch(host)
	if len(matches) == 0 {
		return 0, errors.New("hostname did not match the expected regex pattern")
	}
	i, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

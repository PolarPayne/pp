package pp

import (
	"errors"
	"strings"
	"time"
)

func splitTitle(name string) (time.Time, string, error) {
	var (
		t     time.Time
		title string
		err   error
	)

	split := strings.SplitN(name, " ", 2)
	if len(split) != 2 {
		return t, title, errors.New("")
	}

	t, err = time.Parse("2006-01-02", split[0])
	if err != nil {
		return t, title, errors.New("")
	}

	title = split[1]

	return t, title, nil
}

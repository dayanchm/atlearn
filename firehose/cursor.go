package main

import (
	"os"
	"strconv"
	"strings"
)

const cursorFile = ".cursor"

func SaveCursor(seq int64) error {
	value := strconv.FormatInt(seq, 10)

	return os.WriteFile(
		cursorFile,
		[]byte(value),
		0644,
	)
}

func LoadCursor() (int64, error) {
	data, err := os.ReadFile(cursorFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}

		return 0, err
	}

	return strconv.ParseInt(
		strings.TrimSpace(string(data)),
		10,
		64,
	)
}

package main

import (
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Poetry struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Text   string `json:"text"`
}

// LoadPoetry reads one SQL export or all .sql files below a directory.
// It supports the single-poet MySQL export format in sql/, without a database.
func LoadPoetry(path string) ([]Poetry, error) {
	var poems []Poetry
	err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".sql") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		loaded, err := parsePoetrySQL(string(data))
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		poems = append(poems, loaded...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(poems) == 0 {
		return nil, fmt.Errorf("no poems found in %s", path)
	}
	return poems, nil
}

var poetInsert = regexp.MustCompile(`(?i)INSERT\s+INTO\s+poets\s*\(\s*name\s*\)\s*VALUES\s*\(`)
var poemInsert = regexp.MustCompile(`(?i)INSERT\s+INTO\s+poems\s*\(`)
var poemValues = regexp.MustCompile(`(?i)^\s*poet_id\s*,\s*title\s*,\s*text\s*,\s*source\s*\)\s*VALUES\s*\(\s*@poet_id\s*,`)

func parsePoetrySQL(content string) ([]Poetry, error) {
	match := poetInsert.FindStringIndex(content)
	if match == nil {
		return nil, fmt.Errorf("missing poets insert")
	}
	remaining := content[match[1]:]
	author, err := readSQLString(&remaining)
	if err != nil {
		return nil, fmt.Errorf("poet name: %w", err)
	}
	if strings.TrimSpace(author) == "" {
		return nil, fmt.Errorf("empty poet name")
	}

	var poems []Poetry
	for {
		match = poemInsert.FindStringIndex(remaining)
		if match == nil {
			break
		}
		remaining = remaining[match[1]:]
		values := poemValues.FindStringIndex(remaining)
		if values == nil {
			return nil, fmt.Errorf("poem %d: unsupported insert format", len(poems)+1)
		}
		remaining = remaining[values[1]:]
		fields := make([]string, 3)
		for i := range fields {
			remaining = strings.TrimLeft(remaining, " \t\r\n")
			if i == 2 && strings.HasPrefix(remaining, "NULL") {
				remaining = remaining[4:]
			} else {
				fields[i], err = readSQLString(&remaining)
				if err != nil {
					return nil, fmt.Errorf("poem %d, field %d: %w", len(poems)+1, i+1, err)
				}
			}
			separator := ","
			if i == 2 {
				separator = ")"
			}
			if !consumeSQL(&remaining, separator) {
				return nil, fmt.Errorf("poem %d: expected %q", len(poems)+1, separator)
			}
		}
		if !consumeSQL(&remaining, ";") {
			return nil, fmt.Errorf("poem %d: expected semicolon", len(poems)+1)
		}
		if strings.TrimSpace(fields[0]) == "" || strings.TrimSpace(fields[1]) == "" {
			return nil, fmt.Errorf("poem %d: empty title or text", len(poems)+1)
		}
		poems = append(poems, Poetry{Title: fields[0], Author: author, Text: fields[1]})
	}
	if len(poems) == 0 {
		return nil, fmt.Errorf("no poems found")
	}
	return poems, nil
}

func consumeSQL(input *string, token string) bool {
	*input = strings.TrimLeft(*input, " \t\r\n")
	if !strings.HasPrefix(*input, token) {
		return false
	}
	*input = (*input)[len(token):]
	return true
}

// readSQLString preserves multiline UTF-8 text and decodes MySQL string escapes.
func readSQLString(input *string) (string, error) {
	if !consumeSQL(input, "'") {
		return "", fmt.Errorf("expected quoted SQL string")
	}
	var value strings.Builder
	s := *input
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\'':
			if i+1 < len(s) && s[i+1] == '\'' {
				value.WriteByte('\'')
				i++
				continue
			}
			*input = s[i+1:]
			return value.String(), nil
		case '\\':
			i++
			if i >= len(s) {
				return "", fmt.Errorf("unfinished SQL escape")
			}
			switch s[i] {
			case '0':
				value.WriteByte(0)
			case 'n':
				value.WriteByte('\n')
			case 'r':
				value.WriteByte('\r')
			case 't':
				value.WriteByte('\t')
			case 'b':
				value.WriteByte('\b')
			case 'Z':
				value.WriteByte(26)
			case '%', '_':
				value.WriteByte('\\')
				value.WriteByte(s[i])
			default:
				value.WriteByte(s[i])
			}
		default:
			value.WriteByte(s[i])
		}
	}
	return "", fmt.Errorf("unterminated SQL string")
}

// SelectPoetry chooses a random record; an empty list returns the zero value.
func SelectPoetry(poems []Poetry) Poetry {
	if len(poems) == 0 {
		return Poetry{}
	}
	return poems[rand.IntN(len(poems))]
}

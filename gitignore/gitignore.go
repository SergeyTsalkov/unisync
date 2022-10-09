package gitignore

import (
	"path"
	"strings"
)

var separator = "/"

func Match(pattern, name string, isDir bool) bool {
	pattern = strings.ToLower(pattern)
	name = strings.ToLower(name)

	mustDir := false
	if strings.HasSuffix(pattern, separator) {
		mustDir = true
		pattern = strings.TrimSuffix(pattern, separator)
	}

	anchored := false
	if strings.Contains(pattern, separator) {
		anchored = true
		pattern = strings.TrimPrefix(pattern, separator)
	}

	for {
		patternPart, _, patternLen := parts(pattern)
		namePart, _, nameLen := parts(name)

		if patternLen == 0 || nameLen == 0 {
			break
		}

		match, err := path.Match(patternPart, namePart)
		if err != nil {
			return false
		}
		if match {
			anchored = true

			// if there was a trailing slash in the pattern,
			// the last part of the pattern must match a dir
			if patternLen == 1 && nameLen == 1 && mustDir && !isDir {
				return false
			}

		} else if anchored {
			return false
		}

		name = shift(name)
		if anchored {
			pattern = shift(pattern)
		}
	}

	return len(pattern) == 0
}

func parts(path string) (string, string, int) {
	parts := strings.Split(path, separator)

	if len(parts) == 1 {
		if len(parts[0]) == 0 {
			return "", "", 0
		}
		return parts[0], "", 1
	}

	return parts[0], parts[1], len(parts)
}

func shift(path string) string {
	parts := strings.Split(path, separator)
	return strings.Join(parts[1:], separator)
}

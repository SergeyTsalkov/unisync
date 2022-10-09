package gitignore

import (
	"path"
	"strings"
)

var separator = "/"

func MatchAny(patterns []string, name string, isDir bool) bool {
	for _, pattern := range patterns {
		if Match(pattern, name, isDir) {
			return true
		}
	}
	return false
}

func Match(pattern, name string, isDir bool) bool {
	if len(pattern) == 0 || len(name) == 0 {
		return false
	}

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
		if len(pattern) == 0 || len(name) == 0 {
			break
		}

		patternPart, _ := part(pattern)
		namePart, isLastPartOfName := part(name)

		if isLastPartOfName && mustDir && !isDir {
			return false
		}

		if patternPart == "**" {
			anchored = false
			pattern = shift(pattern)
			continue
		}

		match, err := path.Match(patternPart, namePart)
		if err != nil {
			return false
		}

		if match {
			anchored = true
		}
		if !match && anchored {
			return false
		}

		name = shift(name)
		if anchored {
			pattern = shift(pattern)
		}
	}

	return len(pattern) == 0
}

func part(path string) (string, bool) {
	part, _, hasMore := strings.Cut(path, separator)
	return part, !hasMore
}

func shift(path string) string {
	_, remainder, _ := strings.Cut(path, separator)
	return remainder
}

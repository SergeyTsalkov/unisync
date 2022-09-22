package filelist

import (
	"path/filepath"
	"strings"
)

func clean(list, list2 []*FileListItem) []*FileListItem {
	cleanlist := []*FileListItem{}

	for _, item := range list {
		isClean := true

		for _, item2 := range list2 {
			if HasPrefix(item2.Path, item.Path) {
				isClean = false
				break
			}
		}

		if isClean {
			cleanlist = append(cleanlist, item)
		}
	}

	return cleanlist
}

func HasPrefix(_long, _short string) bool {
	long := strings.Split(_long, string(filepath.Separator))
	short := strings.Split(_short, string(filepath.Separator))

	if len(long) <= len(short) {
		return false
	}

	for len(long) > 0 && len(short) > 0 {
		if long[0] != short[0] {
			return false
		}
		long = long[1:]
		short = short[1:]
	}

	return true
}

package gitignore

import (
	"testing"
)

func TestGitIgnore(t *testing.T) {
	tests := []struct {
		pattern, name string
		isDir         bool
		want          bool
	}{
		// we're not just doing substring matches; "thi" won't match "this"
		{"this", "this/is/mydir", true, true},
		{"is", "this/is/mydir", true, true},
		{"mydir", "this/is/mydir", true, true},
		{"thi", "this/is/mydir", true, false},
		{"this/i", "this/is/mydir", true, false},

		// we can use * and ?
		{"th?s", "this/is/mydir", true, true},
		{"thi**s", "this/is/mydir", true, true},
		{"t?s", "this/is/mydir", true, false},
		{"t*s", "this/is/mydir", true, true},
		{"*", "this/is/mydir", true, true},
		{"??", "this/is/mydir", true, true},
		{"???", "this/is/mydir", true, false},
		{"this*is", "this/is/mydir", true, false},
		{"th?s/?s", "this/is/mydir", true, true},

		// having a / in the middle means we must match path from the start
		{"this/is", "this/is/mydir", true, true},
		{"this/is/mydir", "this/is/mydir", true, true},
		{"this/is/mydirs", "this/is/mydir", true, false},
		{"this/is/mydir/a", "this/is/mydir", true, false},
		{"is/mydir", "this/is/mydir", true, false},

		// having a / at the start means we must match path from the start
		{"/this", "this/is/mydir", true, true},
		{"/is", "this/is/mydir", true, false},

		// having a / at the end means we can only match a directory
		{"this/is/mydir/", "this/is/mydir", true, true},
		{"is/mydir/", "this/is/mydir", true, false},
		{"this/is/", "this/is/mydir", true, true},
		{"this/is.txt/", "this/is.txt", false, false},
		{"this/is.txt/", "this/is.txt", true, true},
		{"this/is/", "this/is/myfile.txt", false, true},
		{"myfile.txt/", "this/is/myfile.txt", false, false},
		{"myfile.txt", "this/is/myfile.txt", false, true},
		{"mydir", "this/is/mydir", true, true},
		{"mydir/", "this/is/mydir", true, true},

		// starting with ** lets us match at any depth
		{"**/is/mydir/", "and/this/is/mydir", true, true},
		{"**/this/is/", "and/this/is/mydir", true, true},
		{"**/this/is", "and/this/is/mydir", true, true},
		{"**/is", "and/this/is/mydir", true, true},
		{"**/mydir", "and/this/is/mydir", true, true},
		{"**/and/this", "and/this/is/mydir", true, true},
		{"**/and", "and/this/is/mydir", true, true},
		{"**/an", "and/this/is/mydir", true, false},
		{"**", "and/this/is/mydir", true, true},

		// putting ** in the middle lets us match 0 or more arbitrary path sections
		{"and/**/this", "and/this/is/mydir", true, true},
		{"and/**/is", "and/this/is/mydir", true, true},
		{"and/**/is/mydir", "and/this/is/mydir", true, true},
		{"and/**/mydir", "and/this/is/mydir", true, true},
		{"and/**/mydir", "and/this/is/mydir", true, true},
		{"an/**/mydir", "and/this/is/mydir", true, false},

		// putting ** at the end is redundant
		{"and/**", "and/this/is/mydir", true, true},
		{"an/**", "and/this/is/mydir", true, false},
		{"this/**", "and/this/is/mydir", true, false},
		{"and/this/**", "and/this/is/mydir", true, true},
	}

	for _, test := range tests {
		result := Match(test.pattern, test.name, test.isDir)
		if result != test.want {
			t.Errorf("%v matches %v (isDir=%v) -> %v (expected %v)", test.pattern, test.name, test.isDir, result, test.want)
		}
	}
}

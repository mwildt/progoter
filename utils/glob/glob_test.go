package glob

import (
	"fmt"
	"testing"
)

func TestGlob(t *testing.T) {

	tests := []struct {
		Pattern string
		Path    string
		Expect  bool
	}{{
		Pattern: "config.file",
		Path:    "config.file",
		Expect:  true,
	}, {
		Pattern: "config.file",
		Path:    "config.file.ext",
	}, {
		Pattern: "config.file",
		Path:    "config",
	}, {
		Pattern: "*",
		Path:    "",
		Expect:  true,
	}, {Pattern: "*",
		Path:   "config",
		Expect: true,
	}, {
		Pattern: "*",
		Path:    "foo/bar",
		Expect:  false,
	}, {
		Pattern: "*/*",
		Path:    "foo/bar",
		Expect:  true,
	}, {
		Pattern: "*/baz/*/*.js",
		Path:    "foo/baz/bar/config.js",
		Expect:  true,
	}, {
		Pattern: "foo.??",
		Path:    "foo.js",
		Expect:  true,
	}, {
		Pattern: "?",
		Path:    "",
		Expect:  false,
	}, {
		Pattern: "?",
		Path:    "A",
		Expect:  true,
	}, {
		Pattern: "?",
		Path:    "/",
		Expect:  true,
	}, {
		Pattern: "**",
		Path:    "",
		Expect:  true,
	}, {
		Pattern: "**",
		Path:    "//",
		Expect:  true,
	}, {
		Pattern: "/foo/**/config.js",
		Path:    "/foo/baz/bar/wtf/config.js",
		Expect:  true,
	}, {
		Pattern: "/foo/**/config.js",
		Path:    "/foo/baz/bar/wtf/file.f",
	}}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s match %s => %t", tt.Pattern, tt.Path, tt.Expect), func(t1 *testing.T) {
			nodes := NewParser(tt.Pattern).parse()
			if match(nodes, tt.Path) != tt.Expect {
				t1.Errorf("match error expect match to be %t", tt.Expect)
			}
		})
	}

}

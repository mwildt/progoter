package tools

import (
	"fmt"
	"testing"
)

func TestFileMatch(t *testing.T) {

	tests := []struct {
		exclusion FileExclusion
		path      string
		expect    bool
	}{{
		exclusion: FileExclusion(".git"),
		path:      ".git/HEAD",
		expect:    true,
	}}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s match %s => %t", tt.exclusion, tt.path, tt.expect), func(t1 *testing.T) {
			if tt.exclusion.Match(tt.path) != tt.expect {
				t1.Errorf("match error expect match to be %t", tt.expect)
			}
		})
	}

}

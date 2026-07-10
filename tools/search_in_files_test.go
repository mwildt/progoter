package tools

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"
)

// TESTKOMMENTAR

func TestGrep(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	res, err := SearchInFilesTool{}.Execute(testDir, `{"pattern":"TESTKOMMENTAR"}`)
	if err != nil {
		t.Fatal(err)
	}
	var searchResult []SearchResult
	json.Unmarshal(res, &searchResult)
	if len(searchResult) != 2 {
		t.Errorf("got %d results, want 2", len(searchResult))
	}

}

func TestGrepRegex(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	res, err := SearchInFilesTool{}.Execute(testDir, `{"pattern":"func[\\s]+errorResponse"}`)
	if err != nil {
		t.Fatal(err)
	}
	var searchResult []SearchResult
	json.Unmarshal(res, &searchResult)
	if len(searchResult) != 1 {
		t.Errorf("got %d results, want 1", len(searchResult))
	} else if searchResult[0].FilePath != "tools.go" {
		t.Errorf("got %q, want tools.go", searchResult[0].FilePath)
	}

}

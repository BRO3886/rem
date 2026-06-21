package rem

import (
	"os"
	"strings"
	"testing"
)

const indexNowKey = "785f0399d4f64b1d1775006ca113f39f"

// TestIndexNowKeyFile guards that the key file name and its single-line contents
// match. If either drifts the key verification step at api.indexnow.org will fail.
func TestIndexNowKeyFile(t *testing.T) {
	path := "website/static/" + indexNowKey + ".txt"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("key file not found at %s: %v", path, err)
	}
	got := strings.TrimSpace(string(data))
	if got != indexNowKey {
		t.Errorf("key file contents %q do not match filename stem %q", got, indexNowKey)
	}
}

// TestIndexNowScriptKey guards that the shell script references the same key.
// The key lives in two places (key file + script); this test catches a drift.
func TestIndexNowScriptKey(t *testing.T) {
	data, err := os.ReadFile("scripts/indexnow.sh")
	if err != nil {
		t.Fatalf("indexnow.sh not found: %v", err)
	}
	needle := `KEY="` + indexNowKey + `"`
	if !strings.Contains(string(data), needle) {
		t.Errorf("scripts/indexnow.sh does not contain %q — key may have drifted", needle)
	}
}

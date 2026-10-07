package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/BRO3886/rem/internal/reminder"
)

func TestJSONParentID(t *testing.T) {
	in := []*reminder.Reminder{
		{ID: "CHILD", Name: "child", ParentID: "PARENT"},
		{ID: "PARENT", Name: "parent"},
	}

	var buf bytes.Buffer
	if err := ExportJSON(&buf, in); err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(buf.String(), `"parent_id": "PARENT"`); got != 1 {
		t.Errorf("expected parent_id on exactly one reminder, got %d:\n%s", got, buf.String())
	}
	if got := strings.Count(buf.String(), `"parent_id"`); got != 1 {
		t.Errorf("top-level reminder should omit parent_id:\n%s", buf.String())
	}
}

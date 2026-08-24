package scrape

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestExtractSLRTimestamps(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
	}{
		{
			name:    "v3 timestamps field",
			payload: `{"timestamps":[{"name":"Intro","timestamp":10},{"name":"Cowgirl","timestamp":120}]}`,
			want:    `[{"Intro":10},{"Cowgirl":120}]`,
		},
		{
			name:    "legacy timeStamps field",
			payload: `{"timeStamps":[{"name":"Intro","ts":10}]}`,
			want:    `[{"Intro":10}]`,
		},
		{
			name:    "v3 key with legacy ts field",
			payload: `{"timestamps":[{"name":"Intro","ts":10}]}`,
			want:    `[{"Intro":10}]`,
		},
		{
			name:    "legacy key with v3 timestamp field",
			payload: `{"timeStamps":[{"name":"Intro","timestamp":10}]}`,
			want:    `[{"Intro":10}]`,
		},
		{
			name:    "entries without a name are skipped",
			payload: `{"timestamps":[{"name":"","timestamp":10},{"timestamp":20},{"name":"  ","timestamp":30},{"name":"Cowgirl","timestamp":120}]}`,
			want:    `[{"Cowgirl":120}]`,
		},
		{
			name:    "entries without a numeric value are skipped",
			payload: `{"timestamps":[{"name":"Intro"},{"name":"Cowgirl","timestamp":120}]}`,
			want:    `[{"Cowgirl":120}]`,
		},
		{
			name:    "missing timestamps key",
			payload: `{"title":"some scene"}`,
			want:    "",
		},
		{
			name:    "non-array timestamps key",
			payload: `{"timestamps":{"name":"Intro"}}`,
			want:    "",
		},
		{
			name:    "empty array",
			payload: `{"timestamps":[]}`,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSLRTimestamps(gjson.Parse(tt.payload), "slr-test")
			if got != tt.want {
				t.Errorf("extractSLRTimestamps() = %q, want %q", got, tt.want)
			}
		})
	}
}

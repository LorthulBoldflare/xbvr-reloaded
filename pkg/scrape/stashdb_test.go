package scrape

import (
	"encoding/json"
	"testing"
)

func TestStashVariablesJSON(t *testing.T) {
	out := StashVariablesJSON(map[string]interface{}{
		"parentStudio": "studio-1",
		"page":         2,
		"title":        `a "quoted" title\with specials`,
	})

	var parsed map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("produced invalid JSON: %v (%s)", err, out)
	}
	input := parsed["input"]
	if input["parentStudio"] != "studio-1" {
		t.Fatalf("parentStudio = %v", input["parentStudio"])
	}
	if input["page"] != float64(2) {
		t.Fatalf("page = %v", input["page"])
	}
	if input["title"] != `a "quoted" title\with specials` {
		t.Fatalf("title not round-tripped: %v", input["title"])
	}
}

func TestGetParentSceneQueryVariable(t *testing.T) {
	withTag := getParentSceneQueryVariable("parent-1", "tag-9", 3, 25)
	var parsed map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(withTag), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v (%s)", err, withTag)
	}
	input := parsed["input"]
	if input["parentStudio"] != "parent-1" {
		t.Fatalf("parentStudio = %v", input["parentStudio"])
	}
	tags, ok := input["tags"].(map[string]interface{})
	if !ok || tags["value"] != "tag-9" || tags["modifier"] != "INCLUDES" {
		t.Fatalf("tags = %v", input["tags"])
	}

	withoutTag := getParentSceneQueryVariable("parent-1", "", 1, 25)
	var parsed2 map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(withoutTag), &parsed2); err != nil {
		t.Fatalf("invalid JSON: %v (%s)", err, withoutTag)
	}
	if _, ok := parsed2["input"]["tags"]; ok {
		t.Fatal("tags key present despite empty tagId")
	}
}

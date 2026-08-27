package config

import "testing"

func TestMigrationStatePublishesProgressAndCompletion(t *testing.T) {
	originalState := State
	originalDispatch := dispatchMigrationWS
	t.Cleanup(func() {
		State = originalState
		dispatchMigrationWS = originalDispatch
	})

	type publication struct {
		topic   string
		message map[string]interface{}
	}
	var publications []publication
	dispatchMigrationWS = func(topic string, message map[string]interface{}) {
		publications = append(publications, publication{topic: topic, message: message})
	}

	UpdateMigrationStatus("0088-indexes", 4, 10, "Adding indexes")
	CompleteMigration()

	if len(publications) != 2 {
		t.Fatalf("expected 2 publications, got %d", len(publications))
	}
	progress := publications[0]
	if progress.topic != migrationStateTopic {
		t.Fatalf("unexpected topic %q", progress.topic)
	}
	if progress.message["is_running"] != true || progress.message["current"] != "0088-indexes" || progress.message["progress"] != 4 || progress.message["total"] != 10 || progress.message["message"] != "Adding indexes" {
		t.Fatalf("unexpected progress payload: %#v", progress.message)
	}

	completed := publications[1]
	if completed.message["is_running"] != false || completed.message["current"] != "" || completed.message["progress"] != 0 || completed.message["total"] != 0 || completed.message["message"] != "" {
		t.Fatalf("unexpected completion payload: %#v", completed.message)
	}
}

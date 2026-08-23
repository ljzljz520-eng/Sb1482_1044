package store

import (
	"path/filepath"
	"testing"

	"example.com/materialconsole/internal/model"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "materials.db")
	storage, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	project := model.NewProject("persist-1", "重开验证", "Aurora One")
	project.Materials = []model.MaterialRecord{{ID: "I-109", ProjectID: project.ID, Kind: "video", Title: "参考片", Status: model.StatusNew}}
	project.Scripts = []model.Script{{ID: "script-1", ProjectID: project.ID, Kind: "host", Title: "口播", Lines: []string{"hello"}}}
	project.Timeline = []model.Timeline{{ID: "scene-1", ProjectID: project.ID, Position: 1, Label: "开场", AssetID: "I-109", Seconds: 8}}
	project.Audit = []model.AuditEvent{{ID: "audit-1", ProjectID: project.ID, Action: "created", Actor: "market", Sequence: 1}}
	project.Attachments = []model.Attachment{{ID: "attachment-1", ProjectID: project.ID, Name: "ref", URI: "ref://one"}}
	project.Workflow = model.Workflow{ID: "workflow-1", ProjectID: project.ID, Name: "launch", Current: model.StageDraft}
	project.Export = model.ExportBundle{ID: "export-1", ProjectID: project.ID, Format: "manifest", Status: model.StatusNew}
	if err := storage.SaveProject(project); err != nil {
		t.Fatal(err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	loaded, err := reopened.GetProject(project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Materials) != 1 || len(loaded.Scripts) != 1 || len(loaded.Timeline) != 1 || len(loaded.Attachments) != 1 {
		t.Fatalf("entities did not survive reopen: %+v", loaded)
	}
	if loaded.Workflow.ID != "workflow-1" || loaded.Export.ID != "export-1" {
		t.Fatalf("lifecycle entities did not survive reopen: %+v", loaded)
	}
	material, err := reopened.GetMaterial("I-109")
	if err != nil || material.Title != "参考片" {
		t.Fatalf("material did not survive reopen: %+v %v", material, err)
	}
}

func TestStoreListsAndDeletesProjects(t *testing.T) {
	storage, err := Open(filepath.Join(t.TempDir(), "list.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	for _, id := range []string{"p-a", "p-b"} {
		if err := storage.SaveProject(model.NewProject(id, id, "product")); err != nil {
			t.Fatal(err)
		}
	}
	items, err := storage.ListProjects()
	if err != nil || len(items) != 2 {
		t.Fatalf("unexpected project list: %d %v", len(items), err)
	}
	if err := storage.DeleteProject("p-a"); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.GetProject("p-a"); err != ErrNotFound {
		t.Fatalf("expected deleted project, got %v", err)
	}
}

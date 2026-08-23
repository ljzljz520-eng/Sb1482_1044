package catalog

import (
	"path/filepath"
	"testing"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/store"
)

func TestCatalogRegistersSearchesAndUpdates(t *testing.T) {
	storage, err := store.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	catalog := New(storage)
	if err := catalog.Register(model.NewProject("catalog-1", "素材台", "Aurora One")); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.AddMaterial("catalog-1", model.MaterialRecord{ID: "I-109", Kind: "video", Title: "主视觉视频", Source: "ref://one", Channel: "video"}); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.AddMaterial("catalog-1", model.MaterialRecord{ID: "I-110", Kind: "copy", Title: "卖点文案", Source: "market", Channel: "store"}); err != nil {
		t.Fatal(err)
	}
	items, err := catalog.Search(model.MaterialQuery{ProjectID: "catalog-1", Text: "主视觉"})
	if err != nil || len(items) != 1 || items[0].ID != "I-109" {
		t.Fatalf("unexpected search: %+v %v", items, err)
	}
	updated, err := catalog.UpdateMaterial("catalog-1", model.MaterialUpdate{RecordID: "I-109", Note: "镜头从左向右", Status: model.StatusReady})
	if err != nil || updated.Note == "" || updated.Status != model.StatusReady {
		t.Fatalf("unexpected update: %+v %v", updated, err)
	}
	if _, err := catalog.SetStage("catalog-1", model.StageReview); err != nil {
		t.Fatal(err)
	}
	if err := catalog.AddAttachment("catalog-1", model.Attachment{ID: "att-1", Name: "参考视频", URI: "ref://one"}); err != nil {
		t.Fatal(err)
	}
	if err := catalog.AddCollaboration("catalog-1", "producer", "请补充片尾字幕"); err != nil {
		t.Fatal(err)
	}
	project, err := catalog.Get("catalog-1")
	if err != nil || len(project.Audit) < 3 || len(project.Attachments) != 1 || len(project.Collaborators) != 1 {
		t.Fatalf("catalog changes missing: %+v %v", project, err)
	}
}

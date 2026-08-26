package timeline

import (
	"path/filepath"
	"testing"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/store"
)

func TestBuilderReordersValidatesAndTotalsScenes(t *testing.T) {
	storage, err := store.Open(filepath.Join(t.TempDir(), "timeline.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	project := model.NewProject("timeline-1", "时间线", "Aurora One")
	if err := storage.SaveProject(project); err != nil {
		t.Fatal(err)
	}
	builder := New(storage)
	for _, item := range []model.Timeline{{ID: "scene-a", Label: "开场", AssetID: "I-109", Seconds: 8}, {ID: "scene-b", Label: "展示", AssetID: "I-110", Seconds: 20}, {ID: "scene-c", Label: "召回", AssetID: "I-111", Seconds: 12}} {
		if _, err := builder.AddScene(project.ID, item); err != nil {
			t.Fatal(err)
		}
	}
	items, err := builder.Reorder(project.ID, []string{"scene-b", "scene-a", "scene-c"})
	if err != nil || items[0].ID != "scene-b" {
		t.Fatalf("unexpected order: %+v %v", items, err)
	}
	if err := builder.Validate(project.ID); err != nil {
		t.Fatal(err)
	}
	total, err := builder.Total(project.ID)
	if err != nil || total != 40 {
		t.Fatalf("unexpected total: %d %v", total, err)
	}
	if Render(items) == "" {
		t.Fatal("timeline render is empty")
	}
}

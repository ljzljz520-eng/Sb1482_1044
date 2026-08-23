package script

import (
	"path/filepath"
	"testing"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/store"
)

func TestDesignerCreatesAndRendersThreeScriptSections(t *testing.T) {
	storage, err := store.Open(filepath.Join(t.TempDir(), "script.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	project := model.NewProject("script-1", "脚本页", "Aurora One")
	project.SellingPoints = []string{"轻量", "续航"}
	project.Channels = []string{"video"}
	if err := storage.SaveProject(project); err != nil {
		t.Fatal(err)
	}
	designer := New(storage)
	items, err := designer.DesignLaunchScripts(project)
	if err != nil || len(items) != 3 {
		t.Fatalf("scripts not designed: %d %v", len(items), err)
	}
	if _, err := designer.Update(project.ID, items[0].ID, []string{"更新后的开场", "一句口号"}, 26); err != nil {
		t.Fatal(err)
	}
	rendered, err := designer.Render(project.ID)
	if err != nil || rendered == "" || !contains(rendered, "更新后的开场") || !contains(rendered, "product-3d") {
		t.Fatalf("unexpected render: %s %v", rendered, err)
	}
}

func contains(value, part string) bool {
	for index := 0; index+len(part) <= len(value); index++ {
		if value[index:index+len(part)] == part {
			return true
		}
	}
	return false
}

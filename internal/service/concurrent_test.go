package service

import (
	"path/filepath"
	"testing"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/store"
)

func TestBusiness009Regression(t *testing.T) {
	storage, err := store.Open(filepath.Join(t.TempDir(), "regression.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	console := New(storage)
	if _, err := console.CreateProject("bug-009", "新品资料", "Aurora One", []string{"轻量"}, []string{"video"}, "即刻出发", "ref://one"); err != nil {
		t.Fatal(err)
	}
	if _, err := console.AddMaterial("bug-009", "I-109", "video", "原始主视觉", "ref://one", "video", "初始备注"); err != nil {
		t.Fatal(err)
	}
	updates := []model.MaterialUpdate{{RecordID: "I-109", Title: "更新后的标题", Status: model.StatusReady}, {RecordID: "I-109", Note: "同步确认", Status: model.StatusReady}}
	if err := console.ApplyConcurrentUpdates("bug-009", updates); err != nil {
		t.Fatal(err)
	}
	material, err := console.catalog.Material("bug-009", "I-109")
	if err != nil {
		t.Fatal(err)
	}
	if material.Title != "更新后的标题" || material.Note != "同步确认" {
		t.Fatalf("concurrent update lost data: %+v", material)
	}
}

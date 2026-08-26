package review

import (
	"path/filepath"
	"testing"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/store"
)

func TestBoardSubmitsApprovesAndArchives(t *testing.T) {
	storage, err := store.Open(filepath.Join(t.TempDir(), "review.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	project := model.NewProject("review-1", "审核流", "Aurora One")
	project.SellingPoints = []string{"轻量"}
	project.Channels = []string{"video"}
	project.Slogan = "即刻出发"
	project.ReferenceVideo = "ref://one"
	project.Materials = []model.MaterialRecord{{ID: "I-109", ProjectID: project.ID, Title: "主视觉", Status: model.StatusReady}}
	project.Scripts = []model.Script{{ID: "host", Kind: "host"}, {ID: "product", Kind: "product-3d"}, {ID: "recall", Kind: "recall"}}
	project.Timeline = []model.Timeline{{ID: "a", Position: 1, AssetID: "I-109", Seconds: 8}, {ID: "b", Position: 2, AssetID: "I-110", Seconds: 12}, {ID: "c", Position: 3, AssetID: "I-111", Seconds: 10}}
	if err := storage.SaveProject(project); err != nil {
		t.Fatal(err)
	}
	board := New(storage)
	if _, err := board.Submit(project.ID, "producer"); err != nil {
		t.Fatal(err)
	}
	if _, err := board.Approve(project.ID, "brand"); err != nil {
		t.Fatal(err)
	}
	if err := board.AddComment(project.ID, "market", "请保留原始卖点"); err != nil {
		t.Fatal(err)
	}
	loaded, err := storage.GetProject(project.ID)
	if err != nil || loaded.Stage != model.StageApproved || !loaded.Materials[0].Approved {
		t.Fatalf("approval not persisted: %+v %v", loaded, err)
	}
	if _, err := board.Archive(project.ID, "producer"); err != nil {
		t.Fatal(err)
	}
	archived, err := storage.GetProject(project.ID)
	if err != nil || archived.Stage != model.StageArchived || !archived.Workflow.Complete {
		t.Fatalf("archive not persisted: %+v %v", archived, err)
	}
}

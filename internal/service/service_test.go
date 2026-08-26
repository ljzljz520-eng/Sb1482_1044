package service

import (
	"path/filepath"
	"testing"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/store"
)

func TestConsoleImportsValidRowsAndReportsIssues(t *testing.T) {
	storage, err := store.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	console := New(storage)
	if _, err := console.CreateProject("service-1", "导入项目", "Aurora One", nil, []string{"video"}, "即刻出发", "ref://one"); err != nil {
		t.Fatal(err)
	}
	accepted, issues, err := console.Import("service-1", []model.ImportRow{{ID: "I-109", Kind: "video", Title: "主视觉", Source: "ref://one"}, {ID: "", Title: "缺编号"}, {ID: "I-109", Title: "重复"}})
	if err != nil || accepted != 1 || len(issues) != 2 {
		t.Fatalf("unexpected import result: %d %v %v", accepted, issues, err)
	}
	items, err := console.SearchMaterials(model.MaterialQuery{ProjectID: "service-1", Text: "主视觉"})
	if err != nil || len(items) != 1 {
		t.Fatalf("unexpected imported materials: %+v %v", items, err)
	}
	if _, err := console.Export("service-1", "manifest"); err == nil {
		t.Fatal("incomplete project was exported")
	}
}

func TestWorkflowCreateReviewArchive(t *testing.T) {
	console, storage := newWorkflowConsole(t, "workflow-1")
	defer storage.Close()
	if _, err := console.AddMaterial("workflow-1", "I-109", "video", "主视觉", "ref://one", "video", "镜头一"); err != nil {
		t.Fatal(err)
	}
	if _, err := console.DesignScripts("workflow-1"); err != nil {
		t.Fatal(err)
	}
	for _, scene := range []struct {
		id, label, asset string
		seconds          int
	}{{"scene-1", "开场", "I-109", 8}, {"scene-2", "三维展示", "I-109", 20}, {"scene-3", "片尾召回", "I-109", 12}} {
		if _, err := console.AddScene("workflow-1", scene.id, scene.label, scene.asset, scene.seconds); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := console.SubmitForReview("workflow-1", "producer"); err != nil {
		t.Fatal(err)
	}
	if _, err := console.Approve("workflow-1", "brand"); err != nil {
		t.Fatal(err)
	}
	project, err := console.Archive("workflow-1", "producer")
	if err != nil || project.Stage != model.StageArchived || !project.Workflow.Complete {
		t.Fatalf("workflow did not archive: %+v %v", project, err)
	}
}

func TestWorkflowSearchUpdatePublish(t *testing.T) {
	console, storage := newWorkflowConsole(t, "workflow-2")
	defer storage.Close()
	if _, err := console.AddMaterial("workflow-2", "I-109", "video", "原始主视觉", "ref://one", "video", "初始备注"); err != nil {
		t.Fatal(err)
	}
	if _, err := console.AddMaterial("workflow-2", "I-110", "copy", "渠道文案", "market", "store", "商品页"); err != nil {
		t.Fatal(err)
	}
	items, err := console.SearchMaterials(model.MaterialQuery{ProjectID: "workflow-2", Text: "主视觉", Channel: "video"})
	if err != nil || len(items) != 1 || items[0].ID != "I-109" {
		t.Fatalf("search did not select I-109: %+v %v", items, err)
	}
	updated, err := console.UpdateMaterial("workflow-2", model.MaterialUpdate{RecordID: "I-109", Title: "确认主视觉", Note: "资料更新已确认", Status: model.StatusReady})
	if err != nil || updated.Title != "确认主视觉" || updated.Note != "资料更新已确认" {
		t.Fatalf("update failed: %+v %v", updated, err)
	}
	project, err := console.Project("workflow-2")
	if err != nil {
		t.Fatal(err)
	}
	if project.Materials[0].Title != "确认主视觉" {
		t.Fatalf("published material incomplete: %+v", project.Materials[0])
	}
}

func TestWorkflowImportReport(t *testing.T) {
	console, storage := newWorkflowConsole(t, "workflow-3")
	defer storage.Close()
	accepted, issues, err := console.Import("workflow-3", []model.ImportRow{{ID: "I-109", Kind: "video", Title: "主视觉", Source: "ref://one", Channel: "video"}, {ID: "I-110", Kind: "copy", Title: "卖点", Source: "market", Channel: "store"}, {ID: "bad", Source: "missing title"}})
	if err != nil || accepted != 2 || len(issues) != 1 {
		t.Fatalf("import validation failed: %d %v %v", accepted, issues, err)
	}
	project, err := console.Project("workflow-3")
	if err != nil {
		t.Fatal(err)
	}
	project.SellingPoints = []string{"轻量"}
	project.Slogan = "即刻出发"
	project.ReferenceVideo = "ref://one"
	project.Scripts = []model.Script{{Kind: "host"}, {Kind: "product-3d"}, {Kind: "recall"}}
	project.Timeline = []model.Timeline{{Seconds: 8}, {Seconds: 12}, {Seconds: 10}}
	if err := storage.SaveProject(project); err != nil {
		t.Fatal(err)
	}
	bundle, err := console.Export("workflow-3", "manifest")
	if err != nil || bundle.Status != model.StatusDone || bundle.Manifest == "" {
		t.Fatalf("report export failed: %+v %v", bundle, err)
	}
}

func newWorkflowConsole(t *testing.T, id string) (*Console, *store.Store) {
	t.Helper()
	storage, err := store.Open(filepath.Join(t.TempDir(), id+".db"))
	if err != nil {
		t.Fatal(err)
	}
	console := New(storage)
	if _, err := console.CreateProject(id, "新品发布素材控制台", "Aurora One", []string{"轻量", "续航"}, []string{"video", "store"}, "即刻出发", "ref://one"); err != nil {
		storage.Close()
		t.Fatal(err)
	}
	return console, storage
}

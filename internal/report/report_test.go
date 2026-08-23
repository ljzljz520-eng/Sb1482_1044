package report

import (
	"testing"

	"example.com/materialconsole/internal/model"
)

func TestReportValidatesAndExportsSummary(t *testing.T) {
	project := model.NewProject("report-1", "导出页", "Aurora One")
	project.SellingPoints = []string{"轻量"}
	project.Channels = []string{"video"}
	project.Slogan = "即刻出发"
	project.ReferenceVideo = "ref://one"
	project.Materials = []model.MaterialRecord{{ID: "I-109"}}
	project.Scripts = []model.Script{{Kind: "host"}, {Kind: "product-3d"}, {Kind: "recall"}}
	project.Timeline = []model.Timeline{{Seconds: 8}, {Seconds: 12}, {Seconds: 10}}
	if missing := Validate(project); len(missing) != 0 {
		t.Fatalf("unexpected missing fields: %v", missing)
	}
	bundle, err := BuildExport(project, "json")
	if err != nil || bundle.Status != model.StatusDone || bundle.Manifest == "" {
		t.Fatalf("export failed: %+v %v", bundle, err)
	}
	if FormatSummary(BuildSummary(project)) == "" || ExportManifest(project) == "" {
		t.Fatal("summary formatting failed")
	}

	project.ReferenceVideo = ""
	if len(Validate(project)) != 1 {
		t.Fatal("missing reference video was not detected")
	}
}

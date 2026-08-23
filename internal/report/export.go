package report

import (
	"fmt"
	"strings"

	"example.com/materialconsole/internal/model"
)

func BuildExport(project model.Project, format string) (model.ExportBundle, error) {
	if format != "json" && format != "manifest" && format != "text" {
		return model.ExportBundle{}, fmt.Errorf("unsupported export format: %s", format)
	}
	if missing := missingActions(project); len(missing) > 0 {
		return model.ExportBundle{}, fmt.Errorf("cannot export incomplete project: %s", strings.Join(missing, ","))
	}
	bundle := model.ExportBundle{ID: project.ID + "-export", ProjectID: project.ID, Format: format, Sections: []string{"project", "materials", "scripts", "timeline", "collaboration", "audit"}, Status: model.StatusDone}
	switch format {
	case "json":
		data, err := ExportJSON(project)
		if err != nil {
			return model.ExportBundle{}, err
		}
		bundle.Manifest = string(data)
	case "manifest":
		bundle.Manifest = ExportManifest(project)
	default:
		bundle.Manifest = FormatSummary(BuildSummary(project))
	}
	return bundle, nil
}

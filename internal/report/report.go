package report

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"example.com/materialconsole/internal/model"
)

type Summary struct {
	ProjectID      string   `json:"project_id"`
	ProjectName    string   `json:"project_name"`
	Stage          string   `json:"stage"`
	MaterialCount  int      `json:"material_count"`
	ScriptCount    int      `json:"script_count"`
	SceneCount     int      `json:"scene_count"`
	Duration       int      `json:"duration"`
	Channels       []string `json:"channels"`
	MissingActions []string `json:"missing_actions"`
}

func BuildSummary(project model.Project) Summary {
	missing := missingActions(project)
	channels := append([]string(nil), project.Channels...)
	sort.Strings(channels)
	return Summary{ProjectID: project.ID, ProjectName: project.Name, Stage: project.Stage, MaterialCount: len(project.Materials), ScriptCount: len(project.Scripts), SceneCount: len(project.Timeline), Duration: project.TotalSeconds(), Channels: channels, MissingActions: missing}
}

func Validate(project model.Project) []string {
	return missingActions(project)
}

func ExportJSON(project model.Project) ([]byte, error) {
	if len(missingActions(project)) > 0 {
		return nil, fmt.Errorf("project is incomplete: %s", strings.Join(missingActions(project), ", "))
	}
	return json.MarshalIndent(BuildSummary(project), "", "  ")
}

func ExportManifest(project model.Project) string {
	sections := []string{"project", "materials", "scripts", "timeline", "collaboration", "audit"}
	return strings.Join([]string{project.ID, project.Name, project.Stage, strings.Join(sections, ",")}, "|")
}

func missingActions(project model.Project) []string {
	missing := []string{}
	if strings.TrimSpace(project.Slogan) == "" {
		missing = append(missing, "slogan")
	}
	if len(project.SellingPoints) == 0 {
		missing = append(missing, "selling_points")
	}
	if len(project.Channels) == 0 {
		missing = append(missing, "channels")
	}
	if project.ReferenceVideo == "" {
		missing = append(missing, "reference_video")
	}
	if len(project.Materials) == 0 {
		missing = append(missing, "materials")
	}
	if len(project.Scripts) < 3 {
		missing = append(missing, "scripts")
	}
	if len(project.Timeline) < 3 {
		missing = append(missing, "timeline")
	}
	return missing
}

func FormatSummary(summary Summary) string {
	return fmt.Sprintf("%s [%s] materials=%d scripts=%d scenes=%d duration=%ds", summary.ProjectName, summary.Stage, summary.MaterialCount, summary.ScriptCount, summary.SceneCount, summary.Duration)
}

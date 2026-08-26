package timeline

import (
	"fmt"
	"sort"
	"strings"

	"example.com/materialconsole/internal/model"
)

type Plan struct {
	Scenes       []model.Timeline `json:"scenes"`
	TotalSeconds int              `json:"total_seconds"`
	OpenSeconds  int              `json:"open_seconds"`
	Middle       int              `json:"middle_seconds"`
	CloseSeconds int              `json:"close_seconds"`
	Warnings     []string         `json:"warnings"`
}

type SceneInput struct {
	ID       string
	Label    string
	AssetID  string
	Seconds  int
	Category string
}

func BuildPlan(inputs []SceneInput) (Plan, error) {
	if len(inputs) < 3 {
		return Plan{}, fmt.Errorf("a launch plan needs open, middle and close scenes")
	}
	plan := Plan{Scenes: make([]model.Timeline, 0, len(inputs)), Warnings: []string{}}
	seen := map[string]bool{}
	for index, input := range inputs {
		if strings.TrimSpace(input.ID) == "" || strings.TrimSpace(input.AssetID) == "" {
			return Plan{}, fmt.Errorf("scene %d needs an id and asset", index+1)
		}
		if seen[input.ID] {
			return Plan{}, fmt.Errorf("scene id is duplicated: %s", input.ID)
		}
		if input.Seconds <= 0 {
			return Plan{}, fmt.Errorf("scene %s must have positive duration", input.ID)
		}
		seen[input.ID] = true
		item := model.Timeline{ID: input.ID, Label: input.Label, AssetID: input.AssetID, Seconds: input.Seconds, Position: index + 1, Status: model.StatusReady}
		plan.Scenes = append(plan.Scenes, item)
		plan.TotalSeconds += input.Seconds
		switch strings.ToLower(input.Category) {
		case "open":
			plan.OpenSeconds += input.Seconds
		case "close":
			plan.CloseSeconds += input.Seconds
		default:
			plan.Middle += input.Seconds
		}
	}
	if plan.TotalSeconds > 180 {
		plan.Warnings = append(plan.Warnings, "launch cut exceeds three minutes")
	}
	if plan.OpenSeconds == 0 {
		plan.Warnings = append(plan.Warnings, "opening scene is not categorized")
	}
	if plan.CloseSeconds == 0 {
		plan.Warnings = append(plan.Warnings, "closing recall scene is not categorized")
	}
	return plan, nil
}

func (b *Builder) ReplacePlan(projectID string, plan Plan) error {
	project, err := b.storage.GetProject(projectID)
	if err != nil {
		return err
	}
	if len(plan.Scenes) < 3 {
		return fmt.Errorf("plan needs at least three scenes")
	}
	project.Timeline = make([]model.Timeline, len(plan.Scenes))
	copy(project.Timeline, plan.Scenes)
	for _, item := range project.Timeline {
		item.ProjectID = projectID
		if err := b.storage.SaveTimeline(item); err != nil {
			return err
		}
	}
	return b.storage.SaveProject(project)
}

func ValidateDurations(items []model.Timeline, minimum, maximum int) []string {
	issues := []string{}
	if minimum < 0 || maximum < minimum {
		return []string{"duration range is invalid"}
	}
	for _, item := range items {
		if item.Seconds < minimum {
			issues = append(issues, item.ID+":too-short")
		}
		if item.Seconds > maximum {
			issues = append(issues, item.ID+":too-long")
		}
	}
	return issues
}

func Compact(items []model.Timeline) []model.Timeline {
	ordered := append([]model.Timeline(nil), items...)
	sort.SliceStable(ordered, func(left, right int) bool {
		if ordered[left].Position == ordered[right].Position {
			return ordered[left].ID < ordered[right].ID
		}
		return ordered[left].Position < ordered[right].Position
	})
	result := make([]model.Timeline, 0, len(ordered))
	for _, item := range ordered {
		if item.Status == model.StatusBlocked {
			continue
		}
		item.Position = len(result) + 1
		result = append(result, item)
	}
	return result
}

func Labels(items []model.Timeline) []string {
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, sceneLabel(item))
	}
	return labels
}

func Assets(items []model.Timeline) []string {
	assets := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		if !seen[item.AssetID] {
			seen[item.AssetID] = true
			assets = append(assets, item.AssetID)
		}
	}
	return assets
}

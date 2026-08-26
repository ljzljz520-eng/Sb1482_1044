package report

import (
	"fmt"
	"sort"
	"strings"

	"example.com/materialconsole/internal/model"
)

type Detail struct {
	Project      model.Project   `json:"project"`
	Checklist    model.Checklist `json:"checklist"`
	MaterialRows []MaterialRow   `json:"material_rows"`
	ScriptRows   []ScriptRow     `json:"script_rows"`
	SceneRows    []SceneRow      `json:"scene_rows"`
	AuditRows    []string        `json:"audit_rows"`
	CommentCount int             `json:"comment_count"`
}

type MaterialRow struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Channel  string `json:"channel"`
	Kind     string `json:"kind"`
	Status   string `json:"status"`
	Approved bool   `json:"approved"`
	Note     string `json:"note"`
}

type ScriptRow struct {
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	Lines    int    `json:"lines"`
	Duration int    `json:"duration"`
	Status   string `json:"status"`
}

type SceneRow struct {
	Position int    `json:"position"`
	Label    string `json:"label"`
	AssetID  string `json:"asset_id"`
	Seconds  int    `json:"seconds"`
}

func BuildDetail(project model.Project) Detail {
	detail := Detail{Project: project, Checklist: project.Checklist(), MaterialRows: []MaterialRow{}, ScriptRows: []ScriptRow{}, SceneRows: []SceneRow{}, AuditRows: AuditLines(project), CommentCount: len(project.Collaborators)}
	for _, material := range project.Materials {
		detail.MaterialRows = append(detail.MaterialRows, MaterialRow{ID: material.ID, Title: material.Title, Channel: material.Channel, Kind: material.Kind, Status: material.Status, Approved: material.Approved, Note: material.Note})
	}
	for _, script := range project.Scripts {
		detail.ScriptRows = append(detail.ScriptRows, ScriptRow{Kind: script.Kind, Title: script.Title, Lines: len(script.Lines), Duration: script.Duration, Status: script.Status})
	}
	for _, scene := range project.Timeline {
		detail.SceneRows = append(detail.SceneRows, SceneRow{Position: scene.Position, Label: scene.Label, AssetID: scene.AssetID, Seconds: scene.Seconds})
	}
	sort.SliceStable(detail.MaterialRows, func(left, right int) bool { return detail.MaterialRows[left].ID < detail.MaterialRows[right].ID })
	sort.SliceStable(detail.SceneRows, func(left, right int) bool { return detail.SceneRows[left].Position < detail.SceneRows[right].Position })
	return detail
}

func RenderDetail(detail Detail) string {
	lines := []string{detail.Project.Name + " / " + detail.Project.Product, "stage=" + detail.Project.Stage, "checklist=" + strings.Join(detail.Checklist.Items, ",")}
	lines = append(lines, renderMaterials(detail.MaterialRows)...)
	lines = append(lines, renderScripts(detail.ScriptRows)...)
	lines = append(lines, renderScenes(detail.SceneRows)...)
	lines = append(lines, detail.AuditRows...)
	return strings.Join(lines, "\n")
}

func renderMaterials(items []MaterialRow) []string {
	lines := []string{"materials:"}
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%s %s %s %s approved=%t note=%s", item.ID, item.Kind, item.Channel, item.Status, item.Approved, item.Note))
	}
	return lines
}

func renderScripts(items []ScriptRow) []string {
	lines := []string{"scripts:"}
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%s %s lines=%d duration=%d status=%s", item.Kind, item.Title, item.Lines, item.Duration, item.Status))
	}
	return lines
}

func renderScenes(items []SceneRow) []string {
	lines := []string{"timeline:"}
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("%d %s %s %ds", item.Position, item.Label, item.AssetID, item.Seconds))
	}
	return lines
}

func CSV(project model.Project) string {
	lines := []string{"id,kind,title,channel,status,approved,note"}
	for _, material := range project.Materials {
		lines = append(lines, strings.Join([]string{csvValue(material.ID), csvValue(material.Kind), csvValue(material.Title), csvValue(material.Channel), csvValue(material.Status), fmt.Sprint(material.Approved), csvValue(material.Note)}, ","))
	}
	return strings.Join(lines, "\n") + "\n"
}

func csvValue(value string) string {
	if strings.ContainsAny(value, ",\"\n") {
		return "\"" + strings.ReplaceAll(value, "\"", "\"\"") + "\""
	}
	return value
}

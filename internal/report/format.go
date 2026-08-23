package report

import (
	"fmt"
	"sort"
	"strings"

	"example.com/materialconsole/internal/model"
)

func AuditLines(project model.Project) []string {
	items := append([]model.AuditEvent(nil), project.Audit...)
	sort.SliceStable(items, func(left, right int) bool { return items[left].Sequence < items[right].Sequence })
	result := make([]string, 0, len(items))
	for _, event := range items {
		result = append(result, fmt.Sprintf("%03d %s %s %s", event.Sequence, event.Action, event.Actor, event.Detail))
	}
	return result
}

func CollaborationDigest(project model.Project) string {
	parts := make([]string, 0, len(project.Collaborators))
	for _, note := range project.Collaborators {
		state := "open"
		if note.Resolved {
			state = "resolved"
		}
		parts = append(parts, note.Author+":"+state+":"+note.Message)
	}
	return strings.Join(parts, "\n")
}

package report

import (
	"sort"
	"strings"

	"example.com/materialconsole/internal/model"
)

func SectionNames(project model.Project) []string {
	sections := []string{"project"}
	if len(project.Materials) > 0 {
		sections = append(sections, "materials")
	}
	if len(project.Scripts) > 0 {
		sections = append(sections, "scripts")
	}
	if len(project.Timeline) > 0 {
		sections = append(sections, "timeline")
	}
	if len(project.Collaborators) > 0 {
		sections = append(sections, "collaboration")
	}
	if len(project.Audit) > 0 {
		sections = append(sections, "audit")
	}
	return sections
}

func MissingAttachmentNames(project model.Project) []string {
	missing := []string{}
	for _, attachment := range project.RequiredAttachments() {
		if strings.TrimSpace(attachment.URI) == "" {
			missing = append(missing, attachment.Name)
		}
	}
	sort.Strings(missing)
	return missing
}

func UnresolvedAudit(project model.Project) []model.AuditEvent {
	items := make([]model.AuditEvent, 0)
	for _, event := range project.Audit {
		if event.Action != "approved" && event.Action != "archived" {
			items = append(items, event)
		}
	}
	sort.SliceStable(items, func(left, right int) bool { return items[left].Sequence < items[right].Sequence })
	return items
}

func ExportReady(project model.Project) bool {
	return len(project.Checklist().Items) == 0 && len(MissingAttachmentNames(project)) == 0 && len(project.OpenComments()) == 0
}

package model

import (
	"fmt"
	"strings"
)

type Checklist struct {
	ProjectIdentity bool     `json:"project_identity"`
	BriefComplete   bool     `json:"brief_complete"`
	MaterialsReady  bool     `json:"materials_ready"`
	ScriptsReady    bool     `json:"scripts_ready"`
	TimelineReady   bool     `json:"timeline_ready"`
	ReviewReady     bool     `json:"review_ready"`
	Items           []string `json:"items"`
}

func (p Project) Checklist() Checklist {
	checklist := Checklist{ProjectIdentity: strings.TrimSpace(p.ID) != "" && strings.TrimSpace(p.Name) != "" && strings.TrimSpace(p.Product) != ""}
	checklist.BriefComplete = strings.TrimSpace(p.Slogan) != "" && strings.TrimSpace(p.ReferenceVideo) != "" && len(p.SellingPoints) > 0 && len(p.Channels) > 0
	checklist.MaterialsReady = materialsReady(p.Materials)
	checklist.ScriptsReady = len(p.Scripts) >= 3 && scriptsHaveContent(p.Scripts)
	checklist.TimelineReady = len(p.Timeline) >= 3 && timelineHasAssets(p.Timeline)
	checklist.ReviewReady = checklist.ProjectIdentity && checklist.BriefComplete && checklist.MaterialsReady && checklist.ScriptsReady && checklist.TimelineReady
	checklist.Items = checklistMissing(checklist)
	return checklist
}

func (p Project) CanAdvance(target string) error {
	if !knownStage(target) {
		return ErrInvalidTransition(p.Stage, target)
	}
	switch target {
	case StageReview:
		if !p.Checklist().ReviewReady {
			return fmt.Errorf("project checklist is incomplete")
		}
	case StageApproved:
		if p.Stage != StageReview {
			return ErrInvalidTransition(p.Stage, target)
		}
	case StageArchived:
		if p.Stage != StageApproved || !p.Workflow.Complete {
			return ErrInvalidTransition(p.Stage, target)
		}
	}
	return nil
}

func (p Project) IsFinal() bool { return p.Stage == StageArchived && p.Archived }

func (p Project) OpenComments() []CollaborationNote {
	comments := make([]CollaborationNote, 0)
	for _, note := range p.Collaborators {
		if !note.Resolved {
			comments = append(comments, note)
		}
	}
	return comments
}

func (p Project) RequiredAttachments() []Attachment {
	required := make([]Attachment, 0)
	for _, attachment := range p.Attachments {
		if attachment.Required {
			required = append(required, attachment)
		}
	}
	return required
}

func (p Project) Attachment(id string) (Attachment, bool) {
	for _, attachment := range p.Attachments {
		if attachment.ID == id {
			return attachment, true
		}
	}
	return Attachment{}, false
}

func (p Project) AuditAction(action string) []AuditEvent {
	items := make([]AuditEvent, 0)
	for _, event := range p.Audit {
		if event.Action == action {
			items = append(items, event)
		}
	}
	return items
}

func scriptsHaveContent(items []Script) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if len(item.Lines) == 0 || item.Duration <= 0 {
			return false
		}
	}
	return true
}

func timelineHasAssets(items []Timeline) bool {
	for index, item := range items {
		if item.AssetID == "" || item.Seconds <= 0 || item.Position != index+1 {
			return false
		}
	}
	return true
}

func materialsReady(items []MaterialRecord) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if item.Status == StatusBlocked || (!item.Approved && item.Status != StatusReady) {
			return false
		}
	}
	return true
}

func checklistMissing(checklist Checklist) []string {
	missing := []string{}
	if !checklist.ProjectIdentity {
		missing = append(missing, "project_identity")
	}
	if !checklist.BriefComplete {
		missing = append(missing, "brief")
	}
	if !checklist.MaterialsReady {
		missing = append(missing, "materials")
	}
	if !checklist.ScriptsReady {
		missing = append(missing, "scripts")
	}
	if !checklist.TimelineReady {
		missing = append(missing, "timeline")
	}
	return missing
}

func knownStage(stage string) bool {
	switch stage {
	case StageDraft, StageReview, StageApproved, StageArchived:
		return true
	default:
		return false
	}
}

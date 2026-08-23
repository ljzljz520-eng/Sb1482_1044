package review

import (
	"fmt"
	"strings"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/store"
)

type Board struct{ storage *store.Store }

func New(storage *store.Store) *Board { return &Board{storage: storage} }

func (b *Board) Submit(projectID, actor string) (model.Project, error) {
	project, err := b.storage.GetProject(projectID)
	if err != nil {
		return model.Project{}, err
	}
	if !project.ReadyForReview() {
		return model.Project{}, fmt.Errorf("project is not ready for review")
	}
	if strings.TrimSpace(actor) == "" {
		return model.Project{}, fmt.Errorf("review actor is required")
	}
	project.Stage = model.StageReview
	project.Workflow.Current = model.StageReview
	project.Workflow.Owner = actor
	project.Audit = append(project.Audit, model.AuditEvent{ID: projectID + "-submitted", ProjectID: projectID, Action: "submitted", Actor: actor, Detail: "review requested", Sequence: len(project.Audit) + 1})
	if err := b.storage.SaveProject(project); err != nil {
		return model.Project{}, err
	}
	return project, nil
}

func (b *Board) Approve(projectID, actor string) (model.Project, error) {
	project, err := b.storage.GetProject(projectID)
	if err != nil {
		return model.Project{}, err
	}
	if project.Stage != model.StageReview {
		return model.Project{}, model.ErrInvalidTransition(project.Stage, model.StageApproved)
	}
	if strings.TrimSpace(actor) == "" {
		return model.Project{}, fmt.Errorf("approver is required")
	}
	project.Stage = model.StageApproved
	project.Workflow.Current = model.StageApproved
	project.Workflow.Owner = actor
	for index := range project.Materials {
		project.Materials[index].Approved = true
		if project.Materials[index].Status == model.StatusNew {
			project.Materials[index].Status = model.StatusReady
		}
	}
	for index := range project.Scripts {
		project.Scripts[index].Status = model.StatusReady
	}
	project.Audit = append(project.Audit, model.AuditEvent{ID: projectID + "-approved", ProjectID: projectID, Action: "approved", Actor: actor, Detail: "release approved", Sequence: len(project.Audit) + 1})
	if err := b.storage.SaveProject(project); err != nil {
		return model.Project{}, err
	}
	return project, nil
}

func (b *Board) Reject(projectID, actor, reason string) (model.Project, error) {
	project, err := b.storage.GetProject(projectID)
	if err != nil {
		return model.Project{}, err
	}
	if project.Stage != model.StageReview {
		return model.Project{}, model.ErrInvalidTransition(project.Stage, model.StageDraft)
	}
	if strings.TrimSpace(actor) == "" || strings.TrimSpace(reason) == "" {
		return model.Project{}, fmt.Errorf("rejector and reason are required")
	}
	project.Stage = model.StageDraft
	project.Workflow.Current = model.StageDraft
	project.Audit = append(project.Audit, model.AuditEvent{ID: projectID + "-rejected", ProjectID: projectID, Action: "rejected", Actor: actor, Detail: reason, Sequence: len(project.Audit) + 1})
	return project, b.storage.SaveProject(project)
}

func (b *Board) Archive(projectID, actor string) (model.Project, error) {
	project, err := b.storage.GetProject(projectID)
	if err != nil {
		return model.Project{}, err
	}
	if project.Stage != model.StageApproved {
		return model.Project{}, model.ErrInvalidTransition(project.Stage, model.StageArchived)
	}
	if strings.TrimSpace(actor) == "" {
		return model.Project{}, fmt.Errorf("archiver is required")
	}
	project.Stage = model.StageArchived
	project.Archived = true
	project.Workflow.Current = model.StageArchived
	project.Workflow.Complete = true
	project.Audit = append(project.Audit, model.AuditEvent{ID: projectID + "-archived", ProjectID: projectID, Action: "archived", Actor: actor, Detail: "release materials archived", Sequence: len(project.Audit) + 1})
	return project, b.storage.SaveProject(project)
}

func (b *Board) AddComment(projectID, author, message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("comment is required")
	}
	project, err := b.storage.GetProject(projectID)
	if err != nil {
		return err
	}
	project.Collaborators = append(project.Collaborators, model.CollaborationNote{ID: fmt.Sprintf("%s-comment-%d", projectID, len(project.Collaborators)+1), ProjectID: projectID, Author: author, Message: message})
	return b.storage.SaveProject(project)
}

func (b *Board) ResolveComment(projectID, noteID string) error {
	project, err := b.storage.GetProject(projectID)
	if err != nil {
		return err
	}
	for index := range project.Collaborators {
		if project.Collaborators[index].ID == noteID {
			project.Collaborators[index].Resolved = true
			return b.storage.SaveProject(project)
		}
	}
	return fmt.Errorf("comment not found: %s", noteID)
}

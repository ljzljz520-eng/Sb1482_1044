package catalog

import (
	"fmt"
	"strings"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/store"
)

type Catalog struct{ storage *store.Store }

func New(storage *store.Store) *Catalog { return &Catalog{storage: storage} }

func (c *Catalog) Register(project model.Project) error {
	if err := project.Valid(); err != nil {
		return err
	}
	if _, err := c.storage.GetProject(project.ID); err == nil {
		return fmt.Errorf("project already exists: %s", project.ID)
	}
	return c.storage.SaveProject(project)
}

func (c *Catalog) Get(projectID string) (model.Project, error) {
	return c.storage.GetProject(projectID)
}

func (c *Catalog) List() ([]model.Project, error) { return c.storage.ListProjects() }

func (c *Catalog) AddMaterial(projectID string, material model.MaterialRecord) (model.MaterialRecord, error) {
	project, err := c.storage.GetProject(projectID)
	if err != nil {
		return model.MaterialRecord{}, err
	}
	if material.ID == "" || material.Title == "" {
		return model.MaterialRecord{}, fmt.Errorf("material id and title are required")
	}
	if project.HasMaterial(material.ID) {
		return model.MaterialRecord{}, fmt.Errorf("material already exists: %s", material.ID)
	}
	material.ProjectID = projectID
	material.Status = defaultStatus(material.Status)
	material.Sequence = len(project.Materials) + 1
	project.Materials = append(project.Materials, material)
	project.Audit = append(project.Audit, model.AuditEvent{ID: projectID + "-material-" + material.ID, ProjectID: projectID, Action: "material-created", Actor: "market", Detail: material.Title, Sequence: len(project.Audit) + 1})
	if err := c.storage.SaveProject(project); err != nil {
		return model.MaterialRecord{}, err
	}
	return material, nil
}

func (c *Catalog) UpdateMaterial(projectID string, update model.MaterialUpdate) (model.MaterialRecord, error) {
	project, err := c.storage.GetProject(projectID)
	if err != nil {
		return model.MaterialRecord{}, err
	}
	for index := range project.Materials {
		if project.Materials[index].ID != update.RecordID {
			continue
		}
		if strings.TrimSpace(update.Title) != "" {
			project.Materials[index].Title = strings.TrimSpace(update.Title)
		}
		if update.Note != "" {
			project.Materials[index].Note = update.Note
		}
		if update.Status != "" {
			project.Materials[index].Status = update.Status
		}
		project.Materials[index].Approved = update.Approved
		project.Audit = append(project.Audit, model.AuditEvent{ID: projectID + "-update-" + update.RecordID + fmt.Sprint(len(project.Audit)), ProjectID: projectID, Action: "material-updated", Actor: "market", Detail: update.RecordID, Sequence: len(project.Audit) + 1})
		if err := c.storage.SaveProject(project); err != nil {
			return model.MaterialRecord{}, err
		}
		return project.Materials[index], nil
	}
	return model.MaterialRecord{}, model.ErrMissingMaterial(update.RecordID)
}

func (c *Catalog) Material(projectID, materialID string) (model.MaterialRecord, error) {
	project, err := c.Get(projectID)
	if err != nil {
		return model.MaterialRecord{}, err
	}
	material, ok := project.Material(materialID)
	if !ok {
		return model.MaterialRecord{}, model.ErrMissingMaterial(materialID)
	}
	return material, nil
}

func (c *Catalog) Search(query model.MaterialQuery) ([]model.MaterialRecord, error) {
	materials, err := c.storage.ListMaterials(query.ProjectID)
	if err != nil {
		return nil, err
	}
	text := strings.ToLower(strings.TrimSpace(query.Text))
	result := make([]model.MaterialRecord, 0, len(materials))
	for _, material := range materials {
		if query.Channel != "" && material.Channel != query.Channel {
			continue
		}
		if query.Kind != "" && material.Kind != query.Kind {
			continue
		}
		if query.Status != "" && material.Status != query.Status {
			continue
		}
		if text != "" && !strings.Contains(strings.ToLower(material.Title+" "+material.Source+" "+material.Note), text) {
			continue
		}
		result = append(result, material)
	}
	return sortMaterials(result), nil
}

func (c *Catalog) SetStage(projectID, stage string) (model.Project, error) {
	project, err := c.Get(projectID)
	if err != nil {
		return model.Project{}, err
	}
	if !validStage(stage) {
		return model.Project{}, model.ErrInvalidTransition(project.Stage, stage)
	}
	if project.Stage == model.StageArchived && stage != model.StageArchived {
		return model.Project{}, model.ErrInvalidTransition(project.Stage, stage)
	}
	project.Stage = stage
	project.Workflow.Current = stage
	project.Audit = append(project.Audit, model.AuditEvent{ID: projectID + "-stage-" + fmt.Sprint(len(project.Audit)), ProjectID: projectID, Action: "stage-changed", Actor: "producer", Detail: stage, Sequence: len(project.Audit) + 1})
	if err := c.storage.SaveProject(project); err != nil {
		return model.Project{}, err
	}
	return project, nil
}

func (c *Catalog) AddAttachment(projectID string, attachment model.Attachment) error {
	project, err := c.Get(projectID)
	if err != nil {
		return err
	}
	if attachment.ID == "" || attachment.URI == "" {
		return fmt.Errorf("attachment id and uri are required")
	}
	attachment.ProjectID = projectID
	for _, existing := range project.Attachments {
		if existing.ID == attachment.ID {
			return fmt.Errorf("attachment already exists: %s", attachment.ID)
		}
	}
	project.Attachments = append(project.Attachments, attachment)
	return c.storage.SaveProject(project)
}

func (c *Catalog) AddCollaboration(projectID, author, message string) error {
	project, err := c.Get(projectID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(author) == "" || strings.TrimSpace(message) == "" {
		return fmt.Errorf("collaboration author and message are required")
	}
	note := model.CollaborationNote{ID: projectID + "-note-" + fmt.Sprint(len(project.Collaborators)+1), ProjectID: projectID, Author: author, Message: message}
	project.Collaborators = append(project.Collaborators, note)
	return c.storage.SaveProject(project)
}

func defaultStatus(value string) string {
	if strings.TrimSpace(value) == "" {
		return model.StatusNew
	}
	return value
}

func validStage(value string) bool {
	switch value {
	case model.StageDraft, model.StageReview, model.StageApproved, model.StageArchived:
		return true
	default:
		return false
	}
}

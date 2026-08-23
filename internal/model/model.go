package model

import "strings"

const (
	StageDraft    = "draft"
	StageReview   = "review"
	StageApproved = "approved"
	StageArchived = "archived"
	StatusNew     = "new"
	StatusReady   = "ready"
	StatusBlocked = "blocked"
	StatusDone    = "done"
)

type Project struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Product        string              `json:"product"`
	SellingPoints  []string            `json:"selling_points"`
	Channels       []string            `json:"channels"`
	Slogan         string              `json:"slogan"`
	ReferenceVideo string              `json:"reference_video"`
	Stage          string              `json:"stage"`
	Materials      []MaterialRecord    `json:"materials"`
	Scripts        []Script            `json:"scripts"`
	Timeline       []Timeline          `json:"timeline"`
	Collaborators  []CollaborationNote `json:"collaborators"`
	Audit          []AuditEvent        `json:"audit"`
	Attachments    []Attachment        `json:"attachments"`
	Workflow       Workflow            `json:"workflow"`
	Export         ExportBundle        `json:"export"`
	Archived       bool                `json:"archived"`
}

type MaterialRecord struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Source    string `json:"source"`
	Channel   string `json:"channel"`
	Status    string `json:"status"`
	Note      string `json:"note"`
	Sequence  int    `json:"sequence"`
	Approved  bool   `json:"approved"`
}

type Script struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	Kind      string   `json:"kind"`
	Title     string   `json:"title"`
	Lines     []string `json:"lines"`
	Duration  int      `json:"duration"`
	Status    string   `json:"status"`
	Version   int      `json:"version"`
}

type Timeline struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Position  int    `json:"position"`
	Label     string `json:"label"`
	AssetID   string `json:"asset_id"`
	Seconds   int    `json:"seconds"`
	Status    string `json:"status"`
}

type CollaborationNote struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Author    string `json:"author"`
	Message   string `json:"message"`
	Resolved  bool   `json:"resolved"`
}

type AuditEvent struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Detail    string `json:"detail"`
	Sequence  int    `json:"sequence"`
}

type Attachment struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	MediaType string `json:"media_type"`
	URI       string `json:"uri"`
	Required  bool   `json:"required"`
}

type Workflow struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Current   string `json:"current"`
	Owner     string `json:"owner"`
	Complete  bool   `json:"complete"`
}

type ExportBundle struct {
	ID        string   `json:"id"`
	ProjectID string   `json:"project_id"`
	Format    string   `json:"format"`
	Sections  []string `json:"sections"`
	Status    string   `json:"status"`
	Manifest  string   `json:"manifest"`
}

type MaterialQuery struct {
	ProjectID string
	Text      string
	Channel   string
	Kind      string
	Status    string
}

type MaterialUpdate struct {
	RecordID string
	Title    string
	Note     string
	Status   string
	Approved bool
}

type ImportRow struct {
	ID      string
	Kind    string
	Title   string
	Source  string
	Channel string
	Note    string
}

func NewProject(id, name, product string) Project {
	return Project{ID: id, Name: name, Product: product, Stage: StageDraft, SellingPoints: []string{}, Channels: []string{}, Materials: []MaterialRecord{}, Scripts: []Script{}, Timeline: []Timeline{}, Collaborators: []CollaborationNote{}, Audit: []AuditEvent{}, Attachments: []Attachment{}, Workflow: Workflow{ID: id + "-workflow", ProjectID: id, Name: "launch", Current: StageDraft, Owner: "market"}, Export: ExportBundle{ID: id + "-export", ProjectID: id, Sections: []string{}}}
}

func (p Project) Valid() error {
	if strings.TrimSpace(p.ID) == "" {
		return ErrInvalidProject("project id is required")
	}
	if strings.TrimSpace(p.Name) == "" {
		return ErrInvalidProject("project name is required")
	}
	if strings.TrimSpace(p.Product) == "" {
		return ErrInvalidProject("product is required")
	}
	if p.Stage != StageDraft && p.Stage != StageReview && p.Stage != StageApproved && p.Stage != StageArchived {
		return ErrInvalidProject("unknown project stage")
	}
	return nil
}

func (p Project) HasMaterial(id string) bool {
	for _, item := range p.Materials {
		if item.ID == id {
			return true
		}
	}
	return false
}

func (p Project) Material(id string) (MaterialRecord, bool) {
	for _, item := range p.Materials {
		if item.ID == id {
			return item, true
		}
	}
	return MaterialRecord{}, false
}

func (p Project) Script(kind string) (Script, bool) {
	for _, item := range p.Scripts {
		if item.Kind == kind {
			return item, true
		}
	}
	return Script{}, false
}

func (p Project) TotalSeconds() int {
	total := 0
	for _, item := range p.Timeline {
		total += item.Seconds
	}
	return total
}

func (p Project) ReadyForReview() bool {
	if len(p.Materials) == 0 || len(p.Scripts) < 3 || len(p.Timeline) < 3 {
		return false
	}
	for _, item := range p.Materials {
		if item.Status == StatusBlocked {
			return false
		}
	}
	return true
}

func (p Project) Clone() Project {
	copyProject := p
	copyProject.SellingPoints = append([]string(nil), p.SellingPoints...)
	copyProject.Channels = append([]string(nil), p.Channels...)
	copyProject.Materials = append([]MaterialRecord(nil), p.Materials...)
	copyProject.Scripts = make([]Script, len(p.Scripts))
	for index, item := range p.Scripts {
		copyProject.Scripts[index] = item
		copyProject.Scripts[index].Lines = append([]string(nil), item.Lines...)
	}
	copyProject.Timeline = append([]Timeline(nil), p.Timeline...)
	copyProject.Collaborators = append([]CollaborationNote(nil), p.Collaborators...)
	copyProject.Audit = append([]AuditEvent(nil), p.Audit...)
	copyProject.Attachments = append([]Attachment(nil), p.Attachments...)
	copyProject.Export.Sections = append([]string(nil), p.Export.Sections...)
	return copyProject
}

type DomainError struct{ Message string }

func (e DomainError) Error() string { return e.Message }

func ErrInvalidProject(message string) error { return DomainError{Message: message} }

func ErrMissingMaterial(id string) error { return DomainError{Message: "material not found: " + id} }

func ErrInvalidTransition(stage, target string) error {
	return DomainError{Message: "cannot move " + stage + " to " + target}
}

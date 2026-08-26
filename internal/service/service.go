package service

import (
	"fmt"
	"strings"
	"sync"

	"example.com/materialconsole/internal/catalog"
	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/report"
	"example.com/materialconsole/internal/review"
	"example.com/materialconsole/internal/script"
	"example.com/materialconsole/internal/store"
	"example.com/materialconsole/internal/timeline"
)

type Console struct {
	storage  *store.Store
	catalog  *catalog.Catalog
	designer *script.Designer
	builder  *timeline.Builder
	review   *review.Board
	mu       sync.RWMutex
}

func New(storage *store.Store) *Console {
	return &Console{storage: storage, catalog: catalog.New(storage), designer: script.New(storage), builder: timeline.New(storage), review: review.New(storage)}
}

func (c *Console) CreateProject(id, name, product string, points, channels []string, slogan, reference string) (model.Project, error) {
	project := model.NewProject(id, name, product)
	project.SellingPoints = append([]string(nil), points...)
	project.Channels = append([]string(nil), channels...)
	project.Slogan = strings.TrimSpace(slogan)
	project.ReferenceVideo = strings.TrimSpace(reference)
	if err := c.catalog.Register(project); err != nil {
		return model.Project{}, err
	}
	return project, nil
}

func (c *Console) Project(id string) (model.Project, error) { return c.catalog.Get(id) }

func (c *Console) ListProjects() ([]model.Project, error) { return c.catalog.List() }

func (c *Console) AddMaterial(projectID, id, kind, title, source, channel, note string) (model.MaterialRecord, error) {
	return c.catalog.AddMaterial(projectID, model.MaterialRecord{ID: id, ProjectID: projectID, Kind: kind, Title: title, Source: source, Channel: channel, Note: note})
}

func (c *Console) SearchMaterials(query model.MaterialQuery) ([]model.MaterialRecord, error) {
	return c.catalog.Search(query)
}

func (c *Console) UpdateMaterial(projectID string, update model.MaterialUpdate) (model.MaterialRecord, error) {
	return c.catalog.UpdateMaterial(projectID, update)
}

func (c *Console) DesignScripts(projectID string) ([]model.Script, error) {
	project, err := c.Project(projectID)
	if err != nil {
		return nil, err
	}
	return c.designer.DesignLaunchScripts(project)
}

func (c *Console) UpdateScript(projectID, kind string, lines []string, duration int) (model.Script, error) {
	scriptItem, err := c.designer.Get(projectID, kind)
	if err != nil {
		return model.Script{}, err
	}
	return c.designer.Update(projectID, scriptItem.ID, lines, duration)
}

func (c *Console) AddScene(projectID, id, label, assetID string, seconds int) (model.Timeline, error) {
	return c.builder.AddScene(projectID, model.Timeline{ID: id, ProjectID: projectID, Label: label, AssetID: assetID, Seconds: seconds})
}

func (c *Console) ReorderTimeline(projectID string, order []string) ([]model.Timeline, error) {
	return c.builder.Reorder(projectID, order)
}

func (c *Console) SubmitForReview(projectID, actor string) (model.Project, error) {
	return c.review.Submit(projectID, actor)
}

func (c *Console) Approve(projectID, actor string) (model.Project, error) {
	return c.review.Approve(projectID, actor)
}

func (c *Console) Reject(projectID, actor, reason string) (model.Project, error) {
	return c.review.Reject(projectID, actor, reason)
}

func (c *Console) Archive(projectID, actor string) (model.Project, error) {
	return c.review.Archive(projectID, actor)
}

func (c *Console) AddComment(projectID, author, message string) error {
	return c.review.AddComment(projectID, author, message)
}

func (c *Console) Export(projectID, format string) (model.ExportBundle, error) {
	project, err := c.Project(projectID)
	if err != nil {
		return model.ExportBundle{}, err
	}
	bundle, err := report.BuildExport(project, format)
	if err != nil {
		return model.ExportBundle{}, err
	}
	if err := c.storage.SaveExport(bundle); err != nil {
		return model.ExportBundle{}, err
	}
	project.Export = bundle
	return bundle, c.storage.SaveProject(project)
}

func (c *Console) Summary(projectID string) (report.Summary, error) {
	project, err := c.Project(projectID)
	if err != nil {
		return report.Summary{}, err
	}
	return report.BuildSummary(project), nil
}

func (c *Console) Import(projectID string, rows []model.ImportRow) (int, []string, error) {
	project, err := c.Project(projectID)
	if err != nil {
		return 0, nil, err
	}
	accepted := 0
	issues := []string{}
	for index, row := range rows {
		if strings.TrimSpace(row.ID) == "" || strings.TrimSpace(row.Title) == "" {
			issues = append(issues, fmt.Sprintf("row %d: id and title required", index+1))
			continue
		}
		if project.HasMaterial(row.ID) {
			issues = append(issues, fmt.Sprintf("row %d: duplicate id %s", index+1, row.ID))
			continue
		}
		project.Materials = append(project.Materials, model.MaterialRecord{ID: row.ID, ProjectID: projectID, Kind: row.Kind, Title: row.Title, Source: row.Source, Channel: row.Channel, Note: row.Note, Status: model.StatusNew, Sequence: len(project.Materials) + 1})
		accepted++
	}
	if accepted == 0 && len(issues) > 0 {
		return 0, issues, fmt.Errorf("no import rows accepted")
	}
	project.Audit = append(project.Audit, model.AuditEvent{ID: projectID + "-import-" + fmt.Sprint(len(project.Audit)), ProjectID: projectID, Action: "materials-imported", Actor: "market", Detail: fmt.Sprint(accepted), Sequence: len(project.Audit) + 1})
	return accepted, issues, c.storage.SaveProject(project)
}

func (c *Console) Validate(projectID string) ([]string, error) {
	project, err := c.Project(projectID)
	if err != nil {
		return nil, err
	}
	return report.Validate(project), nil
}

func (c *Console) TimelineTotal(projectID string) (int, error) {
	return c.builder.Total(projectID)
}

func (c *Console) MaterialStats(projectID string) (catalog.MaterialStats, error) {
	project, err := c.Project(projectID)
	if err != nil {
		return catalog.MaterialStats{}, err
	}
	return catalog.BuildStats(project), nil
}

func (c *Console) MaterialPage(projectID string, offset, limit int) (catalog.Page, error) {
	items, err := c.SearchMaterials(model.MaterialQuery{ProjectID: projectID})
	if err != nil {
		return catalog.Page{}, err
	}
	return catalog.PageMaterials(items, offset, limit)
}

func (c *Console) ReplaceTimelinePlan(projectID string, inputs []timeline.SceneInput) (timeline.Plan, error) {
	plan, err := timeline.BuildPlan(inputs)
	if err != nil {
		return timeline.Plan{}, err
	}
	if err := c.builder.ReplacePlan(projectID, plan); err != nil {
		return timeline.Plan{}, err
	}
	return plan, nil
}

func (c *Console) ScriptQuality(projectID string) ([]script.Quality, error) {
	project, err := c.Project(projectID)
	if err != nil {
		return nil, err
	}
	return script.ValidateScripts(project.Scripts, project.Product), nil
}

func (c *Console) Detail(projectID string) (report.Detail, error) {
	project, err := c.Project(projectID)
	if err != nil {
		return report.Detail{}, err
	}
	return report.BuildDetail(project), nil
}

func (c *Console) MaterialsCSV(projectID string) (string, error) {
	project, err := c.Project(projectID)
	if err != nil {
		return "", err
	}
	return report.CSV(project), nil
}

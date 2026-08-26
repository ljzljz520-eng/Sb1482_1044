package script

import (
	"fmt"
	"strings"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/store"
)

type Designer struct{ storage *store.Store }

func New(storage *store.Store) *Designer { return &Designer{storage: storage} }

func (d *Designer) DesignLaunchScripts(project model.Project) ([]model.Script, error) {
	if err := project.Valid(); err != nil {
		return nil, err
	}
	scripts := []model.Script{
		designHost(project),
		designProduct(project),
		designRecall(project),
	}
	for _, item := range scripts {
		if err := d.storage.SaveScript(item); err != nil {
			return nil, err
		}
	}
	project.Scripts = append([]model.Script(nil), scripts...)
	if err := d.storage.SaveProject(project); err != nil {
		return nil, err
	}
	return scripts, nil
}

func (d *Designer) Update(projectID, scriptID string, lines []string, duration int) (model.Script, error) {
	project, err := d.storage.GetProject(projectID)
	if err != nil {
		return model.Script{}, err
	}
	for index := range project.Scripts {
		if project.Scripts[index].ID != scriptID {
			continue
		}
		if len(lines) == 0 {
			return model.Script{}, fmt.Errorf("script lines are required")
		}
		if duration <= 0 {
			return model.Script{}, fmt.Errorf("script duration must be positive")
		}
		project.Scripts[index].Lines = cleanLines(lines)
		project.Scripts[index].Duration = duration
		project.Scripts[index].Version++
		project.Scripts[index].Status = model.StatusReady
		if err := d.storage.SaveScript(project.Scripts[index]); err != nil {
			return model.Script{}, err
		}
		if err := d.storage.SaveProject(project); err != nil {
			return model.Script{}, err
		}
		return project.Scripts[index], nil
	}
	return model.Script{}, fmt.Errorf("script not found: %s", scriptID)
}

func (d *Designer) Get(projectID, kind string) (model.Script, error) {
	project, err := d.storage.GetProject(projectID)
	if err != nil {
		return model.Script{}, err
	}
	for _, item := range project.Scripts {
		if item.Kind == kind {
			return item, nil
		}
	}
	return model.Script{}, fmt.Errorf("script kind not found: %s", kind)
}

func (d *Designer) Render(projectID string) (string, error) {
	project, err := d.storage.GetProject(projectID)
	if err != nil {
		return "", err
	}
	if len(project.Scripts) == 0 {
		return "", fmt.Errorf("project has no scripts")
	}
	var builder strings.Builder
	for _, item := range project.Scripts {
		builder.WriteString("[" + item.Kind + "]\n")
		for _, line := range item.Lines {
			builder.WriteString(line)
			builder.WriteByte('\n')
		}
	}
	return builder.String(), nil
}

func cleanLines(lines []string) []string {
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		value := strings.TrimSpace(line)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}
	return cleaned
}

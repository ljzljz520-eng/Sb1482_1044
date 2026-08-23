package timeline

import (
	"fmt"
	"sort"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/store"
)

type Builder struct{ storage *store.Store }

func New(storage *store.Store) *Builder { return &Builder{storage: storage} }

func (b *Builder) AddScene(projectID string, item model.Timeline) (model.Timeline, error) {
	project, err := b.storage.GetProject(projectID)
	if err != nil {
		return model.Timeline{}, err
	}
	if item.ID == "" || item.Label == "" || item.Seconds <= 0 {
		return model.Timeline{}, fmt.Errorf("timeline scene requires id, label and positive seconds")
	}
	for _, existing := range project.Timeline {
		if existing.ID == item.ID {
			return model.Timeline{}, fmt.Errorf("timeline scene already exists: %s", item.ID)
		}
	}
	item.ProjectID = projectID
	item.Position = len(project.Timeline) + 1
	item.Status = model.StatusReady
	project.Timeline = append(project.Timeline, item)
	if err := b.storage.SaveTimeline(item); err != nil {
		return model.Timeline{}, err
	}
	if err := b.storage.SaveProject(project); err != nil {
		return model.Timeline{}, err
	}
	return item, nil
}

func (b *Builder) Reorder(projectID string, order []string) ([]model.Timeline, error) {
	project, err := b.storage.GetProject(projectID)
	if err != nil {
		return nil, err
	}
	if len(order) != len(project.Timeline) {
		return nil, fmt.Errorf("timeline order must include every scene")
	}
	byID := make(map[string]model.Timeline, len(project.Timeline))
	for _, item := range project.Timeline {
		byID[item.ID] = item
	}
	result := make([]model.Timeline, 0, len(order))
	seen := make(map[string]bool, len(order))
	for position, id := range order {
		item, ok := byID[id]
		if !ok || seen[id] {
			return nil, fmt.Errorf("invalid timeline scene order: %s", id)
		}
		seen[id] = true
		item.Position = position + 1
		result = append(result, item)
	}
	project.Timeline = result
	for _, item := range result {
		if err := b.storage.SaveTimeline(item); err != nil {
			return nil, err
		}
	}
	if err := b.storage.SaveProject(project); err != nil {
		return nil, err
	}
	return result, nil
}

func (b *Builder) List(projectID string) ([]model.Timeline, error) {
	project, err := b.storage.GetProject(projectID)
	if err != nil {
		return nil, err
	}
	items := append([]model.Timeline(nil), project.Timeline...)
	sort.Slice(items, func(left, right int) bool { return items[left].Position < items[right].Position })
	return items, nil
}

func (b *Builder) Validate(projectID string) error {
	items, err := b.List(projectID)
	if err != nil {
		return err
	}
	if len(items) < 3 {
		return fmt.Errorf("timeline needs at least three scenes")
	}
	for position, item := range items {
		if item.Position != position+1 || item.Seconds <= 0 || item.AssetID == "" {
			return fmt.Errorf("timeline scene %s is incomplete", item.ID)
		}
	}
	return nil
}

func (b *Builder) Total(projectID string) (int, error) {
	items, err := b.List(projectID)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, item := range items {
		total += item.Seconds
	}
	return total, nil
}

func sceneLabel(item model.Timeline) string {
	if item.Label == "" {
		return item.ID
	}
	return item.Label
}

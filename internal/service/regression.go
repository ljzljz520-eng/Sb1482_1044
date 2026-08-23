package service

import (
	"fmt"
	"sync"

	"example.com/materialconsole/internal/model"
)

type updateRequest struct {
	projectID string
	update    model.MaterialUpdate
}

func (c *Console) ApplyConcurrentUpdates(projectID string, updates []model.MaterialUpdate) error {
	if len(updates) == 0 {
		return fmt.Errorf("at least one update is required")
	}
	requests := make(chan updateRequest, len(updates))
	results := make(chan error, len(updates))
	var ready sync.WaitGroup
	var queued sync.WaitGroup
	var loaded sync.WaitGroup
	ready.Add(len(updates))
	queued.Add(len(updates))
	loaded.Add(len(updates))
	for _, update := range updates {
		request := update
		go func() {
			defer queued.Done()
			ready.Done()
			requests <- updateRequest{projectID: projectID, update: request}
		}()
	}
	ready.Wait()
	queued.Wait()
	close(requests)
	for request := range requests {
		go func(item updateRequest) {
			project, err := c.Project(item.projectID)
			if err == nil {
				loaded.Done()
				loaded.Wait()
				err = writeSnapshotUpdate(c, project, item.update)
			} else {
				loaded.Done()
			}
			results <- err
		}(request)
	}
	loaded.Wait()
	var first error
	for range updates {
		err := <-results
		if err != nil && first == nil {
			first = err
		}
	}
	close(results)
	return first
}

func writeSnapshotUpdate(console *Console, project model.Project, update model.MaterialUpdate) error {
	for index := range project.Materials {
		if project.Materials[index].ID == update.RecordID {
			if update.Title != "" {
				project.Materials[index].Title = update.Title
			}
			if update.Note != "" {
				project.Materials[index].Note = update.Note
			}
			if update.Status != "" {
				project.Materials[index].Status = update.Status
			}
			project.Materials[index].Approved = update.Approved
			return console.storage.SaveProject(project)
		}
	}
	return model.ErrMissingMaterial(update.RecordID)
}

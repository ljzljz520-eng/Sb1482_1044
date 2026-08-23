package service

import (
	"fmt"
	"strings"
	"sync"

	"example.com/materialconsole/internal/model"
)

type updateRequest struct {
	projectID string
	update    model.MaterialUpdate
}

// applyUpdate performs a serialized read-modify-write so that concurrent
// updates to the same (or different) materials never overwrite one another.
// Each update re-reads the freshest project from storage, applies only its own
// fields, and persists the result while holding the console-wide lock. This
// guarantees that no field carried by an update is silently lost when another
// update racing against it writes back a stale snapshot.
func (c *Console) applyUpdate(request updateRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	project, err := c.Project(request.projectID)
	if err != nil {
		return err
	}
	return writeSnapshotUpdate(c, project, request.update)
}

// ApplyConcurrentUpdates fans the supplied updates out across goroutines but
// serializes each individual update through applyUpdate. Updates may therefore
// be prepared concurrently, yet every field they carry is durably applied
// against the latest project state and no update overwrites another.
func (c *Console) ApplyConcurrentUpdates(projectID string, updates []model.MaterialUpdate) error {
	if len(updates) == 0 {
		return fmt.Errorf("at least one update is required")
	}
	requests := make(chan updateRequest, len(updates))
	results := make(chan error, len(updates))
	var queued sync.WaitGroup
	queued.Add(len(updates))
	for _, update := range updates {
		request := update
		go func() {
			defer queued.Done()
			requests <- updateRequest{projectID: projectID, update: request}
		}()
	}
	queued.Wait()
	close(requests)
	for request := range requests {
		go func(item updateRequest) {
			results <- c.applyUpdate(item)
		}(request)
	}
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
		// Record the update in the audit log so the material's number stays
		// visible in detail/records. Mirrors Catalog.UpdateMaterial, which
		// the single-update path already emits.
		project.Audit = append(project.Audit, model.AuditEvent{
			ID:        project.ID + "-update-" + update.RecordID + fmt.Sprint(len(project.Audit)),
			ProjectID: project.ID,
			Action:    "material-updated",
			Actor:     "market",
			Detail:    update.RecordID,
			Sequence:  len(project.Audit) + 1,
		})
		return console.storage.SaveProject(project)
	}
	return model.ErrMissingMaterial(update.RecordID)
}

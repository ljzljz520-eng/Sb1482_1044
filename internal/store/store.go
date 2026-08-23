package store

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"example.com/materialconsole/internal/model"
	"go.etcd.io/bbolt"
)

var ErrNotFound = errors.New("stored record not found")

var bucketNames = [][]byte{
	[]byte("projects"), []byte("materials"), []byte("scripts"), []byte("timelines"),
	[]byte("audits"), []byte("attachments"), []byte("workflows"), []byte("exports"),
}

type Store struct {
	db *bbolt.DB
	mu sync.RWMutex
}

func Open(path string) (*Store, error) {
	if filepath.Clean(path) == "." {
		return nil, fmt.Errorf("storage path is required")
	}
	db, err := bbolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open storage: %w", err)
	}
	store := &Store{db: db}
	if err := store.initialize(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range bucketNames {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) SaveProject(project model.Project) error {
	if err := project.Valid(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Update(func(tx *bbolt.Tx) error {
		if err := putValue(tx, "projects", project.ID, project); err != nil {
			return err
		}
		for _, material := range project.Materials {
			if material.ID != "" {
				if err := putValue(tx, "materials", material.ID, material); err != nil {
					return err
				}
			}
		}
		for _, script := range project.Scripts {
			if script.ID != "" {
				if err := putValue(tx, "scripts", script.ID, script); err != nil {
					return err
				}
			}
		}
		for _, timeline := range project.Timeline {
			if timeline.ID != "" {
				if err := putValue(tx, "timelines", timeline.ID, timeline); err != nil {
					return err
				}
			}
		}
		for _, audit := range project.Audit {
			if audit.ID != "" {
				if err := putValue(tx, "audits", audit.ID, audit); err != nil {
					return err
				}
			}
		}
		for _, attachment := range project.Attachments {
			if attachment.ID != "" {
				if err := putValue(tx, "attachments", attachment.ID, attachment); err != nil {
					return err
				}
			}
		}
		if project.Workflow.ID != "" {
			if err := putValue(tx, "workflows", project.Workflow.ID, project.Workflow); err != nil {
				return err
			}
		}
		if project.Export.ID != "" {
			return putValue(tx, "exports", project.Export.ID, project.Export)
		}
		return nil
	})
}

func (s *Store) GetProject(id string) (model.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var project model.Project
	err := s.db.View(func(tx *bbolt.Tx) error {
		return getValue(tx, "projects", id, &project)
	})
	if err != nil {
		return model.Project{}, err
	}
	return project, nil
}

func (s *Store) ListProjects() ([]model.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var projects []model.Project
	err := s.db.View(func(tx *bbolt.Tx) error {
		return forEachValue(tx, "projects", func(data []byte) error {
			var project model.Project
			if err := decode(data, &project); err != nil {
				return err
			}
			projects = append(projects, project)
			return nil
		})
	})
	return projects, err
}

func (s *Store) DeleteProject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.Bucket([]byte("projects")).Delete([]byte(id)); err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) SaveMaterial(material model.MaterialRecord) error {
	return s.saveEntity("materials", material.ID, material)
}

func (s *Store) GetMaterial(id string) (model.MaterialRecord, error) {
	var material model.MaterialRecord
	err := s.getEntity("materials", id, &material)
	return material, err
}

func (s *Store) ListMaterials(projectID string) ([]model.MaterialRecord, error) {
	var materials []model.MaterialRecord
	err := s.listEntities("materials", func(data []byte) error {
		var material model.MaterialRecord
		if err := decode(data, &material); err != nil {
			return err
		}
		if projectID == "" || material.ProjectID == projectID {
			materials = append(materials, material)
		}
		return nil
	})
	return materials, err
}

func (s *Store) SaveScript(script model.Script) error {
	return s.saveEntity("scripts", script.ID, script)
}

func (s *Store) GetScript(id string) (model.Script, error) {
	var script model.Script
	err := s.getEntity("scripts", id, &script)
	return script, err
}

func (s *Store) SaveTimeline(item model.Timeline) error {
	return s.saveEntity("timelines", item.ID, item)
}

func (s *Store) GetTimeline(id string) (model.Timeline, error) {
	var item model.Timeline
	err := s.getEntity("timelines", id, &item)
	return item, err
}

func (s *Store) SaveAuditEvent(event model.AuditEvent) error {
	return s.saveEntity("audits", event.ID, event)
}

func (s *Store) SaveAttachment(attachment model.Attachment) error {
	return s.saveEntity("attachments", attachment.ID, attachment)
}

func (s *Store) SaveWorkflow(workflow model.Workflow) error {
	return s.saveEntity("workflows", workflow.ID, workflow)
}

func (s *Store) SaveExport(bundle model.ExportBundle) error {
	return s.saveEntity("exports", bundle.ID, bundle)
}

func (s *Store) saveEntity(bucket, id string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Update(func(tx *bbolt.Tx) error { return putValue(tx, bucket, id, value) })
}

func (s *Store) getEntity(bucket, id string, target any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.View(func(tx *bbolt.Tx) error { return getValue(tx, bucket, id, target) })
}

func (s *Store) listEntities(bucket string, visitor func([]byte) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.db.View(func(tx *bbolt.Tx) error { return forEachValue(tx, bucket, visitor) })
}

func putValue(tx *bbolt.Tx, bucket, id string, value any) error {
	data, err := encode(value)
	if err != nil {
		return err
	}
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("bucket %s is missing", bucket)
	}
	return b.Put([]byte(id), data)
}

func getValue(tx *bbolt.Tx, bucket, id string, target any) error {
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("bucket %s is missing", bucket)
	}
	data := b.Get([]byte(id))
	if data == nil {
		return ErrNotFound
	}
	return decode(cloneBytes(data), target)
}

func forEachValue(tx *bbolt.Tx, bucket string, visitor func([]byte) error) error {
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return fmt.Errorf("bucket %s is missing", bucket)
	}
	return b.ForEach(func(_, value []byte) error {
		if value == nil {
			return nil
		}
		return visitor(cloneBytes(value))
	})
}

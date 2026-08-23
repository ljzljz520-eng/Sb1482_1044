package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/store"
)

func (s *Server) exportProject(writer http.ResponseWriter, request *http.Request, projectID string) {
	if request.Method != http.MethodPost {
		writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	format := request.URL.Query().Get("format")
	if format == "" {
		format = "manifest"
	}
	bundle, err := s.console.Export(projectID, format)
	if err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, bundle)
}

func (s *Server) reviewProject(writer http.ResponseWriter, request *http.Request, projectID string) {
	if request.Method != http.MethodPost {
		writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	actor := request.URL.Query().Get("actor")
	project, err := s.console.SubmitForReview(projectID, actor)
	if err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, project)
}

func (s *Server) approveProject(writer http.ResponseWriter, request *http.Request, projectID string) {
	if request.Method != http.MethodPost {
		writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	project, err := s.console.Approve(projectID, request.URL.Query().Get("actor"))
	if err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, project)
}

func (s *Server) archiveProject(writer http.ResponseWriter, request *http.Request, projectID string) {
	if request.Method != http.MethodPost {
		writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	project, err := s.console.Archive(projectID, request.URL.Query().Get("actor"))
	if err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, project)
}

func (s *Server) consoleProjectList() ([]model.Project, error) {
	return s.console.ListProjects()
}

func readJSON(request *http.Request, target any) error {
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	return nil
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, map[string]string{"error": message})
}

func writeStoreError(writer http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(writer, http.StatusNotFound, err.Error())
		return
	}
	writeError(writer, http.StatusInternalServerError, err.Error())
}

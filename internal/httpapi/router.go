package httpapi

import (
	"net/http"
	"strings"

	"example.com/materialconsole/internal/model"
	"example.com/materialconsole/internal/service"
)

type Server struct{ console *service.Console }

func New(console *service.Console) *Server { return &Server{console: console} }

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/projects", s.projects)
	mux.HandleFunc("/projects/", s.project)
	return mux
}

func (s *Server) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok", "service": "material-console"})
}

func (s *Server) projects(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		if request.Method == http.MethodPost {
			s.createProject(writer, request)
			return
		}
		writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	items, err := s.consoleProjectList()
	if err != nil {
		writeError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, items)
}

func (s *Server) project(writer http.ResponseWriter, request *http.Request) {
	parts := strings.Split(strings.Trim(request.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[1] == "" {
		writeError(writer, http.StatusBadRequest, "project id is required")
		return
	}
	projectID := parts[1]
	if len(parts) == 2 {
		s.projectDetail(writer, request, projectID)
		return
	}
	s.projectAction(writer, request, projectID, parts[2])
}

func (s *Server) createProject(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		ID             string   `json:"id"`
		Name           string   `json:"name"`
		Product        string   `json:"product"`
		SellingPoints  []string `json:"selling_points"`
		Channels       []string `json:"channels"`
		Slogan         string   `json:"slogan"`
		ReferenceVideo string   `json:"reference_video"`
	}
	if err := readJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	project, err := s.console.CreateProject(input.ID, input.Name, input.Product, input.SellingPoints, input.Channels, input.Slogan, input.ReferenceVideo)
	if err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(writer, http.StatusCreated, project)
}

func (s *Server) projectDetail(writer http.ResponseWriter, request *http.Request, projectID string) {
	if request.Method != http.MethodGet {
		writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	project, err := s.console.Project(projectID)
	if err != nil {
		writeStoreError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, project)
}

func (s *Server) projectAction(writer http.ResponseWriter, request *http.Request, projectID, action string) {
	switch action {
	case "materials":
		s.materials(writer, request, projectID)
	case "summary":
		s.summary(writer, request, projectID)
	case "export":
		s.exportProject(writer, request, projectID)
	case "review":
		s.reviewProject(writer, request, projectID)
	case "approve":
		s.approveProject(writer, request, projectID)
	case "archive":
		s.archiveProject(writer, request, projectID)
	default:
		writeError(writer, http.StatusNotFound, "unknown project action")
	}
}

func (s *Server) materials(writer http.ResponseWriter, request *http.Request, projectID string) {
	if request.Method == http.MethodGet {
		items, err := s.console.SearchMaterials(model.MaterialQuery{ProjectID: projectID, Text: request.URL.Query().Get("q"), Channel: request.URL.Query().Get("channel"), Kind: request.URL.Query().Get("kind")})
		if err != nil {
			writeStoreError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, items)
		return
	}
	if request.Method == http.MethodPost {
		var input model.MaterialRecord
		if err := readJSON(request, &input); err != nil {
			writeError(writer, http.StatusBadRequest, err.Error())
			return
		}
		item, err := s.console.AddMaterial(projectID, input.ID, input.Kind, input.Title, input.Source, input.Channel, input.Note)
		if err != nil {
			writeError(writer, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(writer, http.StatusCreated, item)
		return
	}
	writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *Server) summary(writer http.ResponseWriter, request *http.Request, projectID string) {
	if request.Method != http.MethodGet {
		writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	summary, err := s.console.Summary(projectID)
	if err != nil {
		writeStoreError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, summary)
}

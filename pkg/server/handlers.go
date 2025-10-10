package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"github.com/opencodehq/opencode/pkg/agents"
)

type QueryRequest struct {
	Prompt          string   `json:"prompt"`
	AgentName       string   `json:"agent_name"`
	SystemPrompt    string   `json:"system_prompt"`
	Model           string   `json:"model"`
	AllowedTools    []string `json:"allowed_tools"`
	MaxIterations   int      `json:"max_iterations"`
	EnableStreaming bool     `json:"enable_streaming"`
}

type QueryResponse struct {
	Messages  []*schema.Message `json:"messages"`
	SessionID string            `json:"session_id,omitempty"`
}

type SessionRequest struct {
	AgentName string `json:"agent_name"`
	Prompt    string `json:"prompt"`
}

type SessionMessageRequest struct {
	Message string `json:"message"`
}

type AgentsResponse struct {
	Agents []string `json:"agents"`
}

func (s *Server) registerRoutes() {
	if s.mux == nil {
		s.mux = http.NewServeMux()
	}
	s.mux.HandleFunc("/api/v1/health", s.withJSON(s.health))
	s.mux.HandleFunc("/api/v1/query", s.withJSON(s.query))
	s.mux.HandleFunc("/api/v1/session", s.withJSON(s.sessionRoot))
	s.mux.HandleFunc("/api/v1/session/", s.withJSON(s.sessionRoutes))
	s.mux.HandleFunc("/api/v1/agents", s.withJSON(s.listAgents))
}

type httpHandler func(http.ResponseWriter, *http.Request) error

func (s *Server) withJSON(next httpHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			status := http.StatusInternalServerError
			if serr, ok := err.(interface{ StatusCode() int }); ok {
				status = serr.StatusCode()
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
		}
	}
}

type statusError struct {
	status int
	err    error
}

func (e statusError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e statusError) StatusCode() int {
	if e.status == 0 {
		return http.StatusInternalServerError
	}
	return e.status
}

func newStatusError(status int, err error) error {
	if err == nil {
		return nil
	}
	return statusError{status: status, err: err}
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return newStatusError(http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	return nil
}

func (s *Server) query(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return newStatusError(http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return newStatusError(http.StatusBadRequest, err)
	}
	agentName := strings.TrimSpace(req.AgentName)
	if agentName == "" {
		agentName = "code-writer"
	}
	agent, err := agents.Get(r.Context(), agentName, s.modelMgr, s.toolReg)
	if err != nil {
		return newStatusError(http.StatusNotFound, err)
	}
	messages := make([]*schema.Message, 0)
	if strings.TrimSpace(req.Prompt) != "" {
		messages = append(messages, schema.UserMessage(req.Prompt))
	}
	iter := agent.Run(r.Context(), &adk.AgentInput{Messages: messages, EnableStreaming: req.EnableStreaming})
	collected := collectMessages(iter)
	session := s.sessionMgr.Create(collected)
	writeJSON(w, http.StatusOK, QueryResponse{Messages: collected, SessionID: session.ID})
	return nil
}

func (s *Server) sessionRoot(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return newStatusError(http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
	var req SessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return newStatusError(http.StatusBadRequest, err)
	}
	messages := make([]*schema.Message, 0)
	if strings.TrimSpace(req.Prompt) != "" {
		messages = append(messages, schema.UserMessage(req.Prompt))
	}
	session := s.sessionMgr.Create(messages)
	writeJSON(w, http.StatusCreated, QueryResponse{Messages: session.Messages, SessionID: session.ID})
	return nil
}

func (s *Server) sessionRoutes(w http.ResponseWriter, r *http.Request) error {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/session/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		return newStatusError(http.StatusNotFound, errors.New("session id required"))
	}
	id := parts[0]
	if len(parts) == 1 {
		if r.Method != http.MethodDelete {
			return newStatusError(http.StatusMethodNotAllowed, errors.New("method not allowed"))
		}
		s.sessionMgr.Delete(id)
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	if len(parts) == 2 && parts[1] == "message" {
		if r.Method != http.MethodPost {
			return newStatusError(http.StatusMethodNotAllowed, errors.New("method not allowed"))
		}
		var req SessionMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return newStatusError(http.StatusBadRequest, err)
		}
		session, ok := s.sessionMgr.Append(id, schema.UserMessage(req.Message))
		if !ok {
			return newStatusError(http.StatusNotFound, errors.New("session not found"))
		}
		writeJSON(w, http.StatusOK, QueryResponse{Messages: session.Messages, SessionID: session.ID})
		return nil
	}
	return newStatusError(http.StatusNotFound, errors.New("unknown session route"))
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return newStatusError(http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
	writeJSON(w, http.StatusOK, AgentsResponse{Agents: agents.List()})
	return nil
}

func collectMessages(iter *adk.AsyncIterator[*adk.AgentEvent]) []*schema.Message {
	messages := make([]*schema.Message, 0)
	if iter == nil {
		return messages
	}
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event == nil || event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		msg, err := event.Output.MessageOutput.GetMessage()
		if err != nil {
			continue
		}
		if msg != nil {
			messages = append(messages, msg)
		}
	}
	return messages
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

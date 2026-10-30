package server

import (
	"net/http"

	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/scholar7r/serenia/option"
)

func (s *Server) Hanlder() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /s/{uuid}", s.handleProfile)
	mux.HandleFunc("GET /s/{uuid}/{profile}", s.handleNamedProfile)

	return mux
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	user, ok := s.userMap[uuid]
	if !ok {
		s.logger.Warn("reject request: unknown uuid ", uuid)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if len(user.Profiles) == 0 {
		s.logger.Error("request for uuid ", uuid, ": user has no profile")
		http.Error(w, "user has no profile", http.StatusNotFound)
		return
	}

	s.serveProfile(w, user, user.Profiles[0])
}

func (s *Server) handleNamedProfile(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	user, ok := s.userMap[uuid]
	if !ok {
		s.logger.Warn("reject request: unknown uuid ", uuid)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profileName := r.PathValue("profile")
	if !common.Contains(user.Profiles, profileName) {
		s.logger.Error("request for uuid ", uuid, ": profile not permitted: ", profileName)
		http.Error(w, "profile not permitted: "+profileName, http.StatusForbidden)
		return
	}

	s.serveProfile(w, user, profileName)
}

func (s *Server) serveProfile(w http.ResponseWriter, user option.User, profileName string) {
	p := s.profile.ProfileByName(profileName)
	if p == nil {
		s.logger.Error("request for user ", user.Name, ": profile not found: ", profileName)
		http.Error(w, "profile not found: "+profileName, http.StatusNotFound)
		return
	}

	options, err := p.Render(user.Variables())
	if err != nil {
		s.logger.Error("request for user ", user.Name, ": ", E.Cause(err, "render profile"))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	content, err := json.MarshalContext(s.ctx, options)
	if err != nil {
		s.logger.Error("request for user ", user.Name, ": ", E.Cause(err, "marshal config"))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)

	s.logger.Info("served profile ", profileName, " for user ", user.Name)
}

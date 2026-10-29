package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/sagernet/sing-box/log"
	boxOption "github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/json/badoption"
	"github.com/sagernet/sing/service"
	"github.com/scholar7r/serenia/option"
	"github.com/scholar7r/serenia/profile"
)

const shutdownTimeout = 3 * time.Second

type Server struct {
	createdAt  time.Time
	ctx        context.Context
	logFactory log.Factory
	logger     log.Logger
	httpServer *http.Server
	profile    *profile.Manager
	users      []option.User
	userMap    map[string]option.User
}

func New(ctx context.Context, options option.Options) (*Server, error) {
	ctx = service.ContextWithDefaultRegistry(ctx)
	createdAt := time.Now()
	logFactory, err := log.New(log.Options{
		Context:       ctx,
		Options:       common.PtrValueOrDefault(options.Log),
		DefaultWriter: os.Stderr,
		BaseTime:      createdAt,
	})
	if err != nil {
		return nil, E.Cause(err, "create log factory")
	}

	profileManager, err := profile.NewManager(
		ctx,
		logFactory.NewLogger("profile"),
		common.Map(options.Outbounds, func(it badoption.Listable[boxOption.Outbound]) []boxOption.Outbound {
			return it
		}),
		options.Profiles,
	)
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]option.User)
	for i, u := range options.Users {
		if u.UUID == "" {
			return nil, E.New("initialize user [", i, "]: missing uuid")
		}
		if _, exists := userMap[u.UUID]; exists {
			return nil, E.New("initialize user [", i, "]: duplicate uuid: ", u.UUID)
		}
		userMap[u.UUID] = u
	}

	server := &Server{
		createdAt:  createdAt,
		ctx:        ctx,
		logFactory: logFactory,
		logger:     logFactory.Logger(),
		profile:    profileManager,
		users:      options.Users,
		userMap:    userMap,
	}

	server.httpServer = &http.Server{
		Addr:    options.Listen,
		Handler: server.Hanlder(),
	}

	return server, nil
}

func (s *Server) Start() error {
	err := s.logFactory.Start()
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}

	s.logger.Info("server started at ", listener.Addr())

	go func() {
		err := s.httpServer.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("server serve error: ", err)
		}
	}()

	s.logger.Info("serenity started (", F.Seconds(time.Since(s.createdAt).Seconds()), "s)")

	return nil
}

func (s *Server) Close() error {
	s.logger.Info("serenity shutting down")

	ctx, cancel := context.WithTimeout(s.ctx, shutdownTimeout)
	defer cancel()

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		_ = s.httpServer.Close()
	}

	return E.Errors(err, common.Close(s.logFactory))
}

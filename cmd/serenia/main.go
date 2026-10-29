package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/log"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/scholar7r/serenia/option"
	"github.com/scholar7r/serenia/server"
)

const defaultCfgPath = "/etc/serenia/config.json"

func main() {
	var (
		ctx     = context.Background()
		cfgPath string
	)

	flag.StringVar(
		&cfgPath,
		"c",
		"",
		"set configuration file path",
	)

	flag.Parse()

	if cfgPath == "" {
		cfgPath = defaultCfgPath
	}

	if err := run(ctx, cfgPath); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cfgPath string) error {
	ctx = include.Context(ctx)

	b, err := os.ReadFile(cfgPath)
	if err != nil {
		return E.Cause(err, "read configuration file")
	}

	options, err := json.UnmarshalExtendedContext[option.Options](ctx, b)
	if err != nil {
		return E.Cause(err, "parse configuration")
	}

	srv, err := server.New(ctx, options)
	if err != nil {
		return E.Cause(err, "create server")
	}

	if err = srv.Start(); err != nil {
		return E.Cause(err, "start server")
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	return srv.Close()
}

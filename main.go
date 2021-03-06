package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	configflag "github.com/ncarlier/webhookd/pkg/config/flag"
	"github.com/ncarlier/za/pkg/api"
	"github.com/ncarlier/za/pkg/config"
	"github.com/ncarlier/za/pkg/logger"
	_ "github.com/ncarlier/za/pkg/outputs/all"
	"github.com/ncarlier/za/pkg/server"
	"github.com/ncarlier/za/pkg/version"
)

func main() {
	conf := &config.Flags{}
	configflag.Bind(conf, "ZA")

	flag.Parse()

	if *version.ShowVersion {
		version.Print()
		os.Exit(0)
	}

	level := "info"
	if conf.Debug {
		level = "debug"
	}
	logger.Init(level)

	logger.Debug.Println("starting ZerØ Analytics server...")

	srv := server.NewServer(conf)

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Debug.Println("server is shutting down...")
		api.Shutdown()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error.Fatalf("could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	addr := conf.ListenAddr
	api.Start()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error.Fatalf("could not listen on %s : %v\n", addr, err)
	}

	<-done
	logger.Debug.Println("server stopped")
}

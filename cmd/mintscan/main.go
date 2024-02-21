package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cosmostation/cosmostation-coreum/app"
	"github.com/cosmostation/cosmostation-coreum/mintscan"
	commonhandler "github.com/cosmostation/cosmostation-coreum/mintscan/common"

	"go.uber.org/zap"

	"github.com/gorilla/mux"
)

var (
	Version = "Development"
	Commit  = ""
)

func main() {
	fileBaseName := "mintscan"
	mApp := app.NewApp(fileBaseName)

	cid, err := mApp.Client.RPC.GetNetworkChainID()

	zap.S().Info("connected chain-id : ", cid)

	if cid != mApp.Config.Chain.ChainID {
		panic(err)
	}

	r := mux.NewRouter()
	r = r.PathPrefix("/v1").Subrouter()
	commonhandler.RegisterHandlers(mApp, r)

	sm := &http.Server{
		Addr:         ":" + mApp.Config.Web.Port,
		Handler:      mintscan.Middleware(r),
		ReadTimeout:  time.Duration(mApp.Config.Web.Timeout) * time.Second, // max time to read request from the client
		WriteTimeout: time.Duration(mApp.Config.Web.Timeout) * time.Second, // max time to write response to the client
	}

	// Start the Mintscan API server.
	go func() {
		zap.S().Infof("Server is running on http://localhost:%s", mApp.Config.Web.Port)
		zap.S().Infof("Version: %s | Commit: %s", Version, Commit)

		err := sm.ListenAndServe()
		if err != nil {
			zap.S().Fatal(err)
			os.Exit(1)
		}
	}()

	TrapSignal(sm)
}

// TrapSignal trap the signal from os
func TrapSignal(sm *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGINT,  // interrupt
		syscall.SIGKILL, // kill
		syscall.SIGTERM, // terminate
		syscall.SIGHUP)  // hangup(reload)

	for {
		sig := <-c // Block until a signal is received.
		switch sig {
		case syscall.SIGHUP:
			// cfg.ReloadConfig()
		default:
			terminate(sm, sig)
			break
		}
	}
}

func terminate(sm *http.Server, sig os.Signal) {
	// Gracefully shutdown the server, waiting max 30 seconds for current operations to complete.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sm.Shutdown(ctx)

	zap.S().Infof("Gracefully shutting down the server: %s", sig)
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/lazygo/lazygo/examples/config"
	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/router"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	fmt.Println("Version:", framework.Version)
	fmt.Println("BuildID:", framework.BuildID)
	ptrConfigPath := flag.String("c", "./config.toml", "config path")
	flag.Parse()

	err := config.Init(*ptrConfigPath)
	if err != nil {
		log.Fatalf("[msg: load config file error] [err: %v]", err)
	}

	app := framework.App()
	app.Debug = config.ServerConfig.Debug

	// Start server
	go func() {
		fmt.Println("Listen " + config.ServerConfig.Addr)
		err = router.Init(app).Start(config.ServerConfig.Addr)
		if err != nil && err != http.ErrServerClosed {
			log.Printf("[msg: shutting down the server] [err: %v]", err)
			app.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		log.Printf("[msg: shutting down the server] [err: %v]", err)
		app.Logger.Fatal(err)
	}
}

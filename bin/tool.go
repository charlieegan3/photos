package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/charlieegan3/toolbelt/pkg/tool"
	"github.com/spf13/viper"

	photosWebsiteTool "github.com/charlieegan3/photos/pkg/tool"
)

func main() {
	viper.SetConfigName("config.tool")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	cfg, ok := viper.Get("tools").(map[string]interface{})
	if !ok {
		log.Fatalf("failed to read tools config in map[string]interface{} format")
		os.Exit(1)
	}

	// configure global cancel context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-c:
			cancel()
		}
	}()

	// create a sample toolbelt
	tb := tool.NewBelt()
	tb.SetConfig(cfg)

	wt := photosWebsiteTool.PhotosWebsite{}
	err = tb.AddTool(ctx, &wt)
	if err != nil {
		log.Fatalf("failed to add tool: %v", err)
	}

	port := 3000
	address := "localhost"
	fmt.Printf("Starting server on http://%s:%d\n", address, port)
	tb.RunServer(ctx, address, fmt.Sprintf("%d", port))
}

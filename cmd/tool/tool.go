package main

import (
	"context"
	"log"

	"github.com/charlieegan3/toolbelt/pkg/tool"
	"github.com/spf13/viper"

	photosTool "github.com/charlieegan3/photos/pkg/tool"
)

func main() {

	ctx := context.Background()

	viper.SetConfigName("config.dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	t := &photosTool.PhotosWebsite{}

	tb := tool.NewBelt()
	tb.SetConfig(viper.Get("tools").(map[string]interface{}))
	err = tb.AddTool(ctx, t)
	if err != nil {
		log.Fatalf("failed to add photos tool: %s", err)
	}

	log.Println(
		"running server on http://" +
			viper.GetString("tools.photos.server.address") +
			":" +
			viper.GetString("tools.photos.server.port"),
	)
	tb.RunServer(
		ctx,
		viper.GetString("tools.photos.server.address"),
		viper.GetString("tools.photos.server.port"),
	)
}

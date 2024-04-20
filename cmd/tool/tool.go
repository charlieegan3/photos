package main

import (
	"context"
	"log"
	"os"

	"github.com/charlieegan3/toolbelt/pkg/database"
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

	params := viper.GetStringMapString("database.params")
	connectionString := viper.GetString("database.connectionString")
	db, err := database.Init(connectionString, params, params["dbname"], false)
	if err != nil {
		log.Fatalf("failed to init DB: %s", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	cfg, ok := viper.Get("tools").(map[string]interface{})
	if !ok {
		log.Fatalf("failed to read tools config in map[string]interface{} format")
		os.Exit(1)
	}

	// init the toolbelt, connecting the database, config and external runner
	tb := tool.NewBelt()
	tb.SetConfig(cfg)
	tb.SetDatabase(db)

	t := &photosTool.PhotosWebsite{}
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

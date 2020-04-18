package sync

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/charlieegan3/photos/internal/pkg/proxy"
	"github.com/charlieegan3/photos/internal/types"
	"github.com/spf13/cobra"
)

var source = ""
var output = ""

// CreateSyncCmd initializes the command used by cobra
func CreateSyncCmd() *cobra.Command {
	syncCmd := cobra.Command{
		Use:   "sync",
		Short: "Refreshes and saves profile data",
		Run:   RunSync,
	}

	return &syncCmd
}

// RunSync clones or pulls a repo into the path
func RunSync(cmd *cobra.Command, args []string) {
	resp, err := proxy.GetURLViaProxy("https://www.instagram.com/charlieegan3/?__a=1")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var profile types.Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		log.Fatal(err)
	}

	for _, v := range profile.Graphql.User.EdgeOwnerToTimelineMedia.Edges {
		fmt.Println(v.Node.ID)
	}
}

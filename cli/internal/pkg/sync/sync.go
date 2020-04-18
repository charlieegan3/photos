package sync

import (
	"fmt"
	"os"

	"github.com/charlieegan3/photos/internal/pkg/git"
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
	files, err := git.ListFiles()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, v := range files {
		fmt.Println(v)
	}
	// resp, err := proxy.GetURLViaProxy("https://www.instagram.com/charlieegan3/?__a=1")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(resp)
	// defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	//
	// var profile types.Profile
	// if err := json.Unmarshal(body, &profile); err != nil {
	// 	log.Fatal(err)
	// }
	//
	// for _, v := range profile.Graphql.User.EdgeOwnerToTimelineMedia.Edges {
	// 	fmt.Println(v.Node.ID)
	// }
}

package cmd

import (
	"strings"

	"github.com/mahmoud-emad/nanopods/internal/container"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <container-name> <image[:tag]>",
	Short: "Run a container",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		containerName := args[0]
		imageName, imageTag := parseImage(args[1])

		return runInit(containerName, imageName, imageTag)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runInit(containerName, imageName, imageTag string) error {
	ctr := container.New(containerName, imageName, imageTag)
	return ctr.Run()
}

func parseImage(input string) (string, string) {
	name, tag, found := strings.Cut(input, ":")

	if !found || tag == "" {
		return input, "latest"
	}

	return name, tag
}

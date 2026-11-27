// Package cmd wires up the uinit command line.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nanopods",
	Short: "nanopods is a lightweight container runtime for running isolated workloads.",
	Long:  "Lightweight isolated workloads, containers, process namespaces and resource limits.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

package main

import "github.com/mahmoud-emad/nanopods/cmd"

/*
# Refactor notes:
 1. Use a cli package such as cobra or urfave/cli to handle command-line arguments and commands more effectively.
 2. Implement error handling for invalid commands or arguments.

The exeution of the nanopods:

	nanopods run <container_name> <image_name> <image_tag>
	There are other command options we will provide in the future such as detach, port mapping, etc.
*/
func main() {
	cmd.Execute()
}

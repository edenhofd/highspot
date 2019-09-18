package main

import (
	"fmt"
	"mixtape/service"
	"os"
)

/* notes for problem go here

*/
func main() {

	// don't need the executable so grab 1 onward
	cmdArgs := os.Args[1:]
	if len(cmdArgs) != 3 {
		fmt.Println("Mixtape should be used as follows: ./mixtape.exe <input_file> <changes_file> <output_file>")
		return
	}
	input := cmdArgs[0]
	changes := cmdArgs[1]
	output := cmdArgs[2]

	fmt.Printf("Running Mixtape with:\nInput = %s\nChanges = %s\nOutput = %s\n\n", input, changes, output)

	// let's create the mixtape builder/modifier
	builder, err := service.NewMixtapeBuilder(input)
	if err != nil {
		// something went wrong. print error and quit
		fmt.Printf("\nERROR: %v\n", err)
		return
	}

	// apply update file
	err = builder.ApplyUpdates(changes)
	if err != nil {
		// something went wrong. print error and quit
		fmt.Printf("\nERROR: %v\n", err)
		return
	}

	// dump our updated mixtape data
	err = builder.ExportEntities(output)
	if err != nil {
		// something went wrong. print error and quit
		fmt.Printf("\nERROR: %v\n", err)
		return
	}

	// everything should be good to go if we are here
	fmt.Printf("Congrats! Your modifications have been applied and your output can be found here: %s", output)
}

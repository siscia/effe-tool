package main

import (
	"effe-tool/builder"
	"effe-tool/factory"
	"fmt"
	"github.com/codegangsta/cli"
	"math/rand"
	"os"
	"time"
)

// main
// create scafholding logic.go package
// create new dir
func main() {

	rand.Seed(time.Now().UnixNano())

	// compile := flag.String("compile", "", "Compile the single source file")
	// out := flag.String("out", "", "Path of the executable generated")
	// dir_out := flag.String("dirout", "", "Directory where put your executables.")

	// flag.Parse()

	fmt.Println("Welcome :)")

	app := cli.NewApp()
	app.Name = "effe-tool"
	app.Usage = "Utility to create, build and use effe."

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "out",
			Value: "",
			Usage: "Name of the executable",
		},
		cli.StringFlag{
			Name:  "dirout",
			Value: "out/",
			Usage: "Directory where to save the file",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "new",
			Aliases: []string{"n"},
			Usage:   "create a new empty effe.",
			Action:  factory.CreateNewEffe,
		},
		{
			Name:    "compile",
			Aliases: []string{"c"},
			Usage:   "compile a single file or a whole directory passed as argument.",
			Action:  builder.Compile,
		},
	}

	app.Run(os.Args)

}

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

func main() {

	rand.Seed(time.Now().UnixNano())

	fmt.Println("Welcome :)")

	app := cli.NewApp()
	app.Name = "effe-tool"
	app.Usage = "Utility to create, build and use effe."
	app.Version = "0.1.0"

	app.Commands = []cli.Command{
		{
			Name:    "new",
			Aliases: []string{"n"},
			Usage:   "Create a new empty effe.",
			Action:  factory.CreateNewEffe,
		},
		{
			Name:    "compile",
			Aliases: []string{"c"},
			Usage:   "Compile a single file or a whole directory passed as argument.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dirout",
					Value: "out/",
					Usage: "Directory where to save the executables.",
				},
				cli.StringFlag{
					Name:  "out",
					Value: "",
					Usage: "Custom name to save your executable.",
				},
			},
			Action: builder.Compile,
		},
	}

	fmt.Println(os.Args)
	app.Run(os.Args)
}

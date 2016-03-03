package factory

import (
	"effe-tool/commons"
	"effe-tool/sources"
	"fmt"
	"github.com/codegangsta/cli"
)

func CreateNewEffe(c *cli.Context) {
	filename := c.Args().First()
	if filename == "" {
		fmt.Println("Provide an argument as filemane for the effe.")
		return
	}
	if err := commons.NewFile(filename, sources.Logic); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Successfully created the new effe, path: " + filename)
	}
}

package factory

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/siscia/effe-tool/commons"
	"github.com/siscia/effe-tool/sources"
)

func CreateNewEffe(c *cli.Context) {
	filename := c.Args().First()
	if filename == "" {
		fmt.Println("Provide an argument as filename for the effe.")
		return
	}
	if err := commons.NewFile(filename, sources.Logic); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Successfully created the new effe, path: " + filename)
	}
}

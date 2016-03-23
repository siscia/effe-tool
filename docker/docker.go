package docker

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/siscia/effe-tool/commons"
	"os"
	"os/exec"
	"path/filepath"
)

func logError(path, msg string) {
	fmt.Println("File: " + path + " | " + msg)
}

func dockerFile(path string) string {
	return `
FROM centurylink/ca-certs

ADD exec exec

ENTRYPOINT ["/exec"]
`
}

func dockerifyDirectory(originalPath string, c *cli.Context) {
	log := func(msg string) {
		logError(originalPath, msg)
	}

	walkAndDockerify := func(path string, f os.FileInfo, _ error) error {
		if f.IsDir() {
			return nil
		}
		if f.Mode().IsRegular() {
			fmt.Println()
			err := dockerifyExec(path, c)
			if err != nil {
				log("Error dockerifying the file.")
			}
		}
		return nil
	}
	filepath.Walk(originalPath, walkAndDockerify)
}

func dockerGetCompleteName(path string) string {
	name, version, err := commons.GetNameVersion(path)
	if (err != nil) || (name == "") {
		hash, err := commons.ExecutableHash(path)
		if err != nil {
			name = filepath.Base(path)
			return name
		}
		return hash
	}
	if (name != "") && (version != "") {
		return name + ":" + version
	}
	return filepath.Base(path)
}

func dockerifyExec(path string, c *cli.Context) error {

	log := func(msg string) {
		logError(path, msg)
	}

	// Creating the temporany dir and the whole struct
	dir := os.TempDir() + "/effedocker-" + commons.RandomSuffix()
	if err := os.Mkdir(dir, 0777); err != nil {
		log("Impossible to create temporany dir")
		return err
	}

	// Create the Dockerfile in the directory
	if err := commons.NewFile(dir+"/Dockerfile", dockerFile(path)); err != nil {
		log("Impossible to create the dockerfile in the temporany dir: " + dir)
		return err
	}

	// Move the executable to the directory
	if err := os.Link(path, dir+"/exec"); err != nil {
		log("Impossible to move the file to the temporany directory: " + dir)
		return err
	}

	name := dockerGetCompleteName(path)

	cmd := exec.Command("docker", "build", "-t", name, dir)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log("Problem invoking docker: " + path)
		fmt.Println(err)
		return err
	}

	if err := cmd.Wait(); err != nil {
		log("Problem with docker: " + path)
		fmt.Println(err)
		return err
	}

	log("Everything went good: " + dir)
	return nil
}

func Dockerify(c *cli.Context) {
	path := c.Args().First()
	f, err := os.Lstat(path)
	if err != nil {
		fmt.Println("File: " + path + " | Impossible to open the file, does it exists ?")
		return
	}
	if f.IsDir() {
		dockerifyDirectory(path, c)
	}
	if f.Mode().IsRegular() {
		dockerifyExec(path, c)
	}
}

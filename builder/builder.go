package builder

import (
	"effe-tool/commons"
	"effe-tool/sources"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Doc     string `json:"doc"`
}

func randomSuffix() string {
	return strconv.Itoa(100000 + rand.Intn(1000000))
}

func GetNameVersion(path string) (name, version string, err error) {
	name = ""
	version = ""
	cmd := exec.Command(path, "-info", "True")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return
	}
	if err = cmd.Start(); err != nil {
		fmt.Println(err)
		return
	}
	var i info
	if err = json.NewDecoder(stdout).Decode(&i); err != nil {
		fmt.Println(err)
		return
	}
	if err = cmd.Wait(); err != nil {
		fmt.Println(err)
		return
	}
	name = i.Name
	version = i.Version
	return
}

func CreateFilenameExecutable(name, version string) string {
	return name + "_v" + version
}

func compileSingleFile(source_path string) (string, error) {
	dir := os.TempDir() + "/effebuild-" + randomSuffix()
	if err := os.Mkdir(dir, 0777); err != nil {
		fmt.Println(err)
		return "", err
	}

	dir_effe := dir + "/src/effe"
	if err := os.MkdirAll(dir_effe, 0777); err != nil {
		fmt.Println(err)
		return "", err
	}

	if err := os.Mkdir(dir_effe+"/logic", 0777); err != nil {
		fmt.Println(err)
		return "", err
	}

	if err := os.Link(source_path, dir_effe+"/logic/logic.go"); err != nil {
		fmt.Println(err)
		return "", err
	}

	if err := commons.NewFile(dir_effe+"/effe.go", sources.Core); err != nil {
		fmt.Println("Impossible to create file, exit.")
		fmt.Println(err)
		return "", err
	}

	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", dir+":"+gopath)
	defer os.Setenv("GOPATH", gopath)

	cmd := exec.Command("go", "build", "-a", "-o", dir+"/out", "-buildmode=exe", dir_effe+"/effe.go")

	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
		return "", err
	}
	return dir + "/out", nil
}

func compileDirectory(path string, c *cli.Context) {}
func compileFile(path string, c *cli.Context) error {
	fmt.Println(*c)
	fmt.Println(path)
	tmpExecPath, err := compileSingleFile(path)
	if err != nil {
		fmt.Println("Impossible to compile: " + path)
		return err
	}
	dirName := c.String("dirout")
	fmt.Println(c.IsSet("dirout"))
	execName := c.String("out")
	execVersion := ""
	if execName == "" {
		execName, execVersion, err = GetNameVersion(tmpExecPath)
	}
	execName = CreateFilenameExecutable(execName, execVersion)
	totalPath, err := filepath.Abs(dirName + `/` + execName)
	fmt.Println(totalPath)
	if err != nil {
		fmt.Println("Error in getting the path.")
		return err
	}
	if err := os.MkdirAll(filepath.Dir(totalPath), 0777); err != nil {
		fmt.Println("Impossible to create the directory: " + filepath.Dir(totalPath))
		return err
	}
	if err := os.Rename(tmpExecPath, totalPath); err != nil {
		fmt.Println(err)
		fmt.Println("Impossible to move the executable, actual path is: " + tmpExecPath)
	}
	return nil

}

func moveExecutable(path string) {}

func Compile(c *cli.Context) {
	path := c.Args().First()
	f, err := os.Lstat(path)
	if err != nil {
		fmt.Println("Impossible to open the file, are you sure it exist ?")
		return
	}
	if f.IsDir() {
		compileDirectory(path, c)
	}
	if f.Mode().IsRegular() {
		err := compileFile(path, c)
		if err != nil {
			fmt.Println(err)
		}
	}
}

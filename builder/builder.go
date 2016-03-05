package builder

import (
	"effe-tool/commons"
	"effe-tool/sources"
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"hash/fnv"
	"io"
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

// randomSuffix return a string to be used to generate random
// temporany directories.
func randomSuffix() string {
	return strconv.Itoa(100000 + rand.Intn(1000000))
}

// getNameVersion execute the binary with the `-info` option.
// Then it parse the standard output, parse the JSON and
// return a name and a version string
func getNameVersion(path string) (name, version string, err error) {
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

func createFilenameExecutable(name, version string) string {
	return name + "_v" + version
}

// executableHash given the path of the executable
// generate an hash to be used as name.
// It is the default way to handle not correct info variable.
func executableHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("File: " + path + " | Error in opening the file to generate hash.")
		return "", err
	}
	defer file.Close()

	hash := fnv.New64a()
	_, err = io.Copy(hash, file)

	if err != nil {
		fmt.Println("File: " + path + " | Error in generating the hash.")
		return "", err
	}

	return strconv.FormatUint(hash.Sum64(), 10), nil
}

// compileSingleFile compile an effe to a single binary.
// It start by creating a temporany directory where it moves
// the logic of the effe, and the core.
// Then it adds the temporany dir just created to the GOPATH
// Finally it invoke the go tool to actually compile the file,
// it redirects the Stdout and the Stderr so that the user can
// actually see compilation errors.
// It returns the path where the executable is been created
func compileSingleFile(sourcePath string) (string, error) {

	// Creating temporany directory and structure
	dir := os.TempDir() + "/effebuild-" + randomSuffix()
	if err := os.Mkdir(dir, 0777); err != nil {
		fmt.Println(err)
		return "", err
	}

	dirEffe := dir + "/src/effe"
	if err := os.MkdirAll(dirEffe, 0777); err != nil {
		fmt.Println(err)
		return "", err
	}

	if err := os.Mkdir(dirEffe+"/logic", 0777); err != nil {
		fmt.Println(err)
		return "", err
	}

	if err := os.Link(sourcePath, dirEffe+"/logic/logic.go"); err != nil {
		fmt.Println(err)
		return "", err
	}

	if err := commons.NewFile(dirEffe+"/effe.go", sources.Core); err != nil {
		fmt.Println("Impossible to create file, exit.")
		fmt.Println(err)
		return "", err
	}

	// adding temporany dir to the GOPATH
	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", dir+":"+gopath)
	defer os.Setenv("GOPATH", gopath)

	// actually compile
	cmd := exec.Command("go", "build", "-a", "-o", dir+"/out", "-buildmode=exe", dirEffe+"/effe.go")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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

// compileFile is the entry point to compile an effe source.
// The actual compilation is done by `compileSingleFile` but
// `compileFile` takes care of move the binary where the user
// is expecting.
// It first compile the file (passed as path).
// It execute the just compiled file to gather information
// about name and version.
// Finally it moves the executable.
//
// `path` is where the effe source is located.
// `dirName` rappresent in which directory save the executable,
// it has a default value set on the flag to `out`.
// `execName` is the name of the executable, if not given
// `compileFile` try to use the effe convetion to provide a name.
func compileFile(path, dirName, execName string) error {
	// Actually compiling
	tmpExecPath, err := compileSingleFile(path)
	if err != nil {
		fmt.Println("File: " + path + " | Impossible to compile.")
		return err
	}

	// Gathering information
	execVersion := ""
	if execName == "" {

		// the user want didn't provide a name for the executable
		// we need to come out with a name

		execName, execVersion, err = getNameVersion(tmpExecPath)
		if err == nil {

			// everything went well and the names comes from the info variable

			execName = createFilenameExecutable(execName, execVersion)
		} else {

			// the info variable doesn't provide the right information
			// we will use the hash of the executable as name

			fmt.Println("File: " + path + " | Error in the executable info.")
			fmt.Println("File: " + path + " | Falling back to use the hash as name.")
			hashName, err := executableHash(tmpExecPath)
			if err != nil {

				// problem generating the hash, we don't move the executable

				fmt.Println("File: " + path + " | Error in generating the hash name.")
				fmt.Println("File: " + path + " | Actual path is: " + tmpExecPath)
				return err
			}
			execName = hashName
		}
	}

	// Moving the file
	totalPath, err := filepath.Abs(dirName + `/` + execName)
	if err != nil {
		fmt.Println("File: " + path + " | Error in getting the absolute path.\nActual path is: " + tmpExecPath)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(totalPath), 0777); err != nil {
		fmt.Println("File: " + path + " | Impossible to create the directory: " + filepath.Dir(totalPath))
		fmt.Println("Actual path is: " + tmpExecPath)
		return err
	}
	if err := os.Rename(tmpExecPath, totalPath); err != nil {
		fmt.Println(err)
		fmt.Println("File: " + path + " | Impossible to move the executable.\nActual path is: " + tmpExecPath)
	}
	fmt.Println("File: " + path + " | Everything went good, the file is been compiled.\nExecutable path: " + totalPath)
	return nil
}

// compileDirectory simply walks the filesystem and
// try to compile every file it find.
// The real job is done by `walkAndCompile`
// walkAndCompile simply does nothing to the directory.
// walkAndCompile preserve the shape of the source dir
// into the executable directory.
func compileDirectory(originalPath string, c *cli.Context) {
	walkAndCompile := func(path string, f os.FileInfo, _ error) error {
		if f.IsDir() {
			return nil
		}
		if f.Mode().IsRegular() {
			fmt.Println()
			relativePath, err := filepath.Rel(originalPath, path)
			if err != nil {
				fmt.Println("File: " + path + " | Error with the relative path.")
				return nil
			}
			execLocation := filepath.Dir(relativePath)
			err = compileFile(path, c.String("dirout")+"/"+execLocation, "")
			if err != nil {
				fmt.Println(err)
			}
		}
		return nil
	}
	filepath.Walk(originalPath, walkAndCompile)
}

// Compile is the main entry point
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
		err := compileFile(path, c.String("dirout"), c.String("out"))
		if err != nil {
			fmt.Println(err)
		}
	}
}

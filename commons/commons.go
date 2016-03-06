package commons

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
)

// NewFile create a new file, the path of the file will be the
// first argument, while its content will be the second one.
func NewFile(path string, source string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = ioutil.WriteFile(path, []byte(source), 0644)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("The file already exist")
}

// randomSuffix return a string to be used to generate random
// temporany directories.
func RandomSuffix() string {
	return strconv.Itoa(100000 + rand.Intn(1000000))
}

type info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Doc     string `json:"doc"`
}

// getNameVersion execute the binary with the `-info` option.
// Then it parse the standard output, parse the JSON and
// return a name and a version string
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

// executableHash given the path of the executable
// generate an hash to be used as name.
// It is the default way to handle not correct info variable.
func ExecutableHash(path string) (string, error) {
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

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Doc     string `json:"doc"`
}

var logic = `
package logic

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"math/rand"
	"net/http"
	"time"
)

var Info string = ` + "`" + `
{
	"name": "hello_effe",
	"version": "0.1",
	"doc" : "Getting start with effe"
}
` + "`" + `

type Context struct {
	value int64
}

func Init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func Start() Context {
	fmt.Println("Start new Context")
	return Context{1 + rand.Int63n(2)}
}

func Run(ctx Context, w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintf(w, "Hello from Effe with logs:  %d", ctx.value)
	log.WithFields(log.Fields{
		"animal": "walrus",
	}).Info("A walrus appears")
	return nil
}

func Stop(ctx Context) { return }
`

var core = `
package main

import (
	"effe/logic"
	"flag"
	"fmt"
	"log/syslog"
	"net/http"
	"sync"
)

func generateHandler(pool *sync.Pool, logger *syslog.Writer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := pool.Get().(logic.Context)
		defer func() {
			if r := recover(); r != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logger.Crit("Logic Panicked")
			}
		}()
		err := logic.Run(ctx, w, r)
		if err != nil {
			logger.Debug(err.Error())
		}
		pool.Put(ctx)
	}
}

func main() {
	port := flag.Int("port", 8080, "Port where serve the effe.")
	info := flag.Bool("info", false, "Print the effe information, then exit.")
	flag.Parse()
	if *info {
		fmt.Println(logic.Info)
		return
	}
	url := fmt.Sprintf(":%d", *port)
	logic.Init()
	logger, _ := syslog.New(syslog.LOG_ERR|syslog.LOG_USER, "Logs From Effe ")
	var ctxPool = &sync.Pool{New: func() interface{} {
		return logic.Start()
	}}
	http.HandleFunc("/", generateHandler(ctxPool, logger))
	http.ListenAndServe(url, nil)
}
`

func random_suffix() string {
	return strconv.Itoa(100000 + rand.Intn(1000000))
}

func new_file(path string, source string) error {
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

func compile_file(source_path string) (string, error) {
	dir := "/tmp/effebuild-" + random_suffix()
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

	if err := new_file(dir_effe+"/effe.go", core); err != nil {
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

func get_name_version(path string) (name, version string, err error) {
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

func create_filename_executable(name, version string) string {
	return name + "_v" + version
}

// main
// create scafholding logic.go package
// create new dir
func main() {

	rand.Seed(time.Now().UnixNano())

	nw := flag.String("new", "", "Create new scafholding file")
	// project := flag.String("project", "", "Create new project")
	compile := flag.String("compile", "", "Compile the single source file")
	out := flag.String("out", "", "Path of the executable generated")
	dir_out := flag.String("dirout", "", "Directory where put your executables.")
	// release := flag.String("release", "", "Compile the files of the directories into binaries")
	// indipendent := flag.Bool("indipendent", true, "Compile all the binaries down to completely indipendent executables")

	flag.Parse()

	fmt.Println("Welcome :)")

	if *nw != "" {
		err := new_file(*nw, logic)
		if err != nil {
			fmt.Println("Error creating the new file.")
			fmt.Println(err)
			return
		} else {
			fmt.Println("New file sucessfully create, path: " + *nw)
		}
	}

	if *compile != "" {
		if compiled_path, err := compile_file(*compile); err != nil {
			fmt.Println(err)
			fmt.Println(compiled_path)
			return
		} else {
			source_name := *out
			if source_name == "" {
				name, version, err := get_name_version(compiled_path)
				if (err != nil) || (name == "") || (version == "") {
					// if no out si defined the executable will be created on the workdir with the same name of the source file
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(`To keep naming coherent inside the project ` +
							`it is suggested to provide the information about both the name and ` +
							`the versione in the logic source code.`)
					}
					source_name = filepath.Base(*compile)
					source_name = strings.Split(source_name, ".")[0]
				} else {
					fmt.Println("Name: " + name)
					fmt.Println("Version: " + version)
					source_name = create_filename_executable(name, version)
				}
			}
			if *dir_out != "" {
				source_name = *dir_out + "/" + source_name
			}
			source_name, err := filepath.Abs(source_name)
			if err != nil {
				// no idea how it could come out an error, no idea why we should care... Still...
				fmt.Println(err)
			}
			if err := os.MkdirAll(filepath.Dir(source_name), 0777); err != nil {
				fmt.Println(err)
			}
			if err := os.Rename(compiled_path, source_name); err != nil {
				fmt.Println(err)
				fmt.Println("It wasn't possible to move the executable, its path is: " + compiled_path)
				return
			} else {
				fmt.Println("The executable is located in: " + source_name)
				return
			}
		}
	}
}

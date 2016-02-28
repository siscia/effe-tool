package main

import (
    "flag"
    "fmt"
    "os"
    "io/ioutil"
    "errors"
    "math/rand"
    "time"
    "strconv"
    "os/exec"
)


var logic = `
   package logic

    import (
        "net/http"
        "fmt"
        "math/rand"
        "time"
    )
    
    type Context struct{
        value int64
    }
    
    func Init() {
        rand.Seed(time.Now().UTC().UnixNano())
    }
    
    func Start() Context {
        fmt.Println("Start new Context")
        return Context{1 + rand.Int63n(2)};
    }
    
    func Run(ctx Context, w http.ResponseWriter, r *http.Request) error {
        fmt.Fprintf(w, "Hello from Logic: %d", ctx.value)
        return nil
    }
    
    func Stop(ctx Context) {return }`


var core = `
    package main
    
    import(
        "effe/logic"
        "net/http"
        "sync"
        "log/syslog"
        "flag"
        "fmt"
    )
    
    func generateHandler(pool *sync.Pool, logger *syslog.Writer) func(http.ResponseWriter, *http.Request) {
        return func(w http.ResponseWriter, r *http.Request){
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
        flag.Parse()
        url := fmt.Sprintf(":%d", *port)
        logic.Init()
        logger, _ := syslog.New(syslog.LOG_ERR | syslog.LOG_USER, "Logs From Effe ")
        var ctxPool = &sync.Pool{New: func () interface{} {
    	return logic.Start()} }
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


// main 
// create scafholding logic.go package
// create new dir
func main() {

    rand.Seed(time.Now().UnixNano())

    nw := flag.String("new", "", "Create new scafholding file")
   // project := flag.String("project", "", "Create new project")
    compile := flag.String("compile", "", "Compile the single source file")
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
	dir := "/tmp/effebuild-" + random_suffix()
	fmt.Println(dir)
	err := os.Mkdir(dir, 0777)
	if err != nil {fmt.Println(err); return}
	err = os.Mkdir(dir + "/logic", 0777)
	if err != nil {fmt.Println(err); return}
	err = os.Link(*compile, dir + "/logic/logic.go")
	if err != nil {fmt.Println(err); return}
	err = new_file(dir + "/effe.go", core)
	if err != nil {
	    fmt.Println("Impossible to create file, exit.")
	    fmt.Println(err)
	    return
	}
	cmd := exec.Command("go", "build", "-pkgdir", dir, "-o", dir + "/out", "-buildmode=exe", dir + "/effe.go")
	err = cmd.Start()
	if err != nil {fmt.Println(err); return}
	err = cmd.Wait()
	if err != nil {fmt.Println(err); return}
    }

}


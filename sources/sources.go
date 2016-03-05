package sources

var Logic = `
package logic

import (
	"fmt"
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
	fmt.Fprintf(w, "Hello from Effe:  %d\n", ctx.value)
	return nil
}

func Stop(ctx Context) { return }
`

var Core = `
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

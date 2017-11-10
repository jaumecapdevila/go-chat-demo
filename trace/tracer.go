package trace

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/jaumecapdevila/chat/config"
)

// Tracer interface
type Tracer interface {
	Trace(...interface{})
}

type tracer struct {
	out io.Writer
}

type logTracer struct{}

type nilTracer struct{}

func (t *tracer) Trace(a ...interface{}) {
	fmt.Fprint(t.out, a...)
	fmt.Fprintln(t.out)
}

func (t *logTracer) Trace(a ...interface{}) {
	log.Print(a...)
}

func (t *nilTracer) Trace(a ...interface{}) {}

// New returns a new Tracer object
func New(w io.Writer) Tracer {
	return &tracer{out: w}
}

// Log returns a Log Tracer object
func Log() Tracer {
	return &logTracer{}
}

// Off returns a silent tracer
func Off() Tracer {
	return &nilTracer{}
}

func init() {
	config := config.Load()
	if _, err := os.Stat(config.Logger.Dir); os.IsNotExist(err) {
		err := os.MkdirAll(config.Logger.Dir, 0755)
		if err != nil {
			fmt.Printf("Error %v", err)
			os.Exit(1)
		}
	}
	if _, err := os.Stat(path.Join(config.Logger.Dir, "error.log")); os.IsNotExist(err) {
		nf, err := os.Create(path.Join(config.Logger.Dir, "error.log"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		log.SetOutput(nf)
		return
	}
	f, err := os.OpenFile(path.Join(config.Logger.Dir, "error.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.SetOutput(f)
}

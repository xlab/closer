// Package closer ensures a clean exit for your Go app.
//
// The aim of this package is to provide an universal way to catch the event of application’s exit
// and perform some actions before it’s too late. Closer doesn’t care about the way application
// tries to exit, i.e. was that a panic or just a signal from the OS, it calls the provided methods
// for cleanup and that’s the whole point.
//
// Exit codes
//
// All errors and panics will be logged if the logging option of `closer.Checked` was set true,
// also the exit code (for `os.Exit`) will be determined accordingly:
//
//   Event         | Default exit code
//   ------------- | -------------
//   error = nil   | 0 (success)
//   error != nil  | 1 (failure)
//   panic         | 1 (failure)
//
package closer

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	// DebugSignalSet is a predefined list of signals to watch for. Usually
	// these signals will terminate the app without executing the code in defer blocks.
	DebugSignalSet = []os.Signal{
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
	}
	// DefaultSignalSet will have syscall.SIGABRT that should be
	// opted out if user wants to debug the stacktrace.
	DefaultSignalSet = append(DebugSignalSet, syscall.SIGABRT)
)

var (
	// ExitCodeOK is a successfull exit code.
	ExitCodeOK = 0
	// ExitCodeErr is a failure exit code.
	ExitCodeErr = 1
	// ExitSignals is the active list of signals to watch for.
	ExitSignals = DefaultSignalSet
)

// Config should be used with Init function to override the defaults.
type Config struct {
	ExitCodeOK  int
	ExitCodeErr int
	ExitSignals []os.Signal
}

var c = newCloser()

type closer struct {
	codeOK     int
	codeErr    int
	signals    []os.Signal
	sem        sync.Mutex
	cleanups   []func()
	errChan    chan struct{}
	doneChan   chan struct{}
	signalChan chan os.Signal
	closeChan  chan struct{}
	holdChan   chan struct{}
	//
	cancelWaitChan chan struct{}
}

func newCloser() *closer {
	c := &closer{
		codeOK:  ExitCodeOK,
		codeErr: ExitCodeErr,
		signals: ExitSignals,
		//
		errChan:    make(chan struct{}),
		doneChan:   make(chan struct{}),
		signalChan: make(chan os.Signal, 1),
		closeChan:  make(chan struct{}),
		holdChan:   make(chan struct{}),
		//
		cancelWaitChan: make(chan struct{}),
	}

	signal.Notify(c.signalChan, c.signals...)

	// start waiting
	go c.wait()
	return c
}

func (c *closer) wait() {
	exitCode := c.codeOK

	// wait for a close request
	select {
	case <-c.cancelWaitChan:
		return
	case <-c.signalChan:
	case <-c.closeChan:
		break
	case <-c.errChan:
		exitCode = c.codeErr
	}

	// ensure we'll exit
	defer os.Exit(exitCode)

	c.sem.Lock()
	defer c.sem.Unlock()
	for _, fn := range c.cleanups {
		fn()
	}
	// done!
	close(c.doneChan)
}

// Close sends a close request.
// The app will be terminated by OS as soon as the first close request will be handled by closer, this
// function will return no sooner. The exit code will always be 0 (success).
func Close() {
	// check if there was a panic
	if x := recover(); x != nil {
		log.Printf("run time panic: %v", x)
		// close with an error
		close(c.errChan)
	} else {
		// normal close
		close(c.closeChan)
	}
	<-c.doneChan
}

func (c *closer) closeErr() {
	close(c.errChan)
	<-c.doneChan
}

// Init allows user to override the defaults (a set of OS signals to watch for, for example).
func Init(cfg Config) {
	c.sem.Lock()
	signal.Stop(c.signalChan)
	close(c.cancelWaitChan)
	c.codeOK = cfg.ExitCodeOK
	c.codeErr = cfg.ExitCodeErr
	c.signals = cfg.ExitSignals
	signal.Notify(c.signalChan, c.signals...)
	go c.wait()
	c.sem.Unlock()
}

// Bind will register the cleanup function that will be called when closer will get a close request.
// All the callbacks will be called in the reverse order they were bound, that's similar to how `defer` works.
func Bind(cleanup func()) {
	c.sem.Lock()
	// store in the reverse order
	s := make([]func(), 0, 1+len(c.cleanups))
	s = append(s, cleanup)
	c.cleanups = append(s, c.cleanups...)
	c.sem.Unlock()
}

// Checked runs the target function and checks for panics and errors it may yield. In case of panic or error, closer
// will terminate the app with an error code, but either case it will call all the bound callbacks beforehand.
// One can use this instead of `defer` if you need to care about errors and panics that always may happen.
// This function optionally can emit log messages via standard `log` package.
func Checked(target func() error, logging bool) {
	defer func() {
		// check if there was a panic
		if x := recover(); x != nil {
			if logging {
				log.Printf("run time panic: %v", x)
			}
			// close with an error
			c.closeErr()
		}
	}()
	if err := target(); err != nil {
		if logging {
			log.Println("error:", err)
		}
		// close with an error
		c.closeErr()
	}
}

// Hold is a helper that may be used to hold the main from returning,
// until the closer will do a proper exit via `os.Exit`.
func Hold() {
	<-c.holdChan
}

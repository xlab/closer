## Closer [![Circle CI](https://circleci.com/gh/xlab/closer/tree/master.svg?style=svg)](https://circleci.com/gh/xlab/closer/tree/master) [![GoDoc](https://godoc.org/github.com/xlab/closer?status.svg)](https://godoc.org/github.com/xlab/closer)

The aim of this package is to provide an universal way to catch the event of application’s exit and perform some actions before it’s too late. `closer` doesn’t care about the way application tries to exit, i.e. was that a panic or just a signal from the OS, it calls the provided methods for cleanup and that’s the whole point.

![demo](https://habrastorage.org/getpro/habr/post_images/f2c/025/0cb/f2c0250cbc4e8519d706b5a35374d40d.png)

### Usage

Be careful, this package is using the singleton pattern (like `net/http` does) and doesn't require any initialisation step. However, there’s an option to provide a custom configuration struct.

```go
// Init allows user to override the defaults (a set of OS signals to watch for, for example).
func Init(cfg Config)

// Close sends a close request.
// The app will be terminated by OS as soon as the first close request will be handled by closer, this
// function will return no sooner. The exit code will always be 0 (success).
func Close()

// Bind will register the cleanup function that will be called when closer will get a close request.
// All the callbacks will be called in the reverse order they were bound, that's similar to how `defer` works.
func Bind(cleanup func())

// Checked runs the target function and checks for panics and errors it may yield. In case of panic or error, closer
// will terminate the app with an error code, but either case it will call all the bound callbacks beforehand.
// One can use this instead of `defer` if you need to care about errors and panics that always may happen.
// This function optionally can emit log messages via standard `log` package.
func Checked(target func() error, logging bool)

// Hold is a helper that may be used to hold the main from returning,
// until the closer will do a proper exit via `os.Exit`.
func Hold()
```

The the usage examples: [example](/cmd/example/main.go), [example-error](/cmd/example-error/main.go) and [example-panic](/cmd/example-panic/main.go).

### Table of exit codes

All errors and panics will be logged if the logging option of `closer.Checked` was set true, also the exit code (for `os.Exit`) will be determined accordingly:

Event         | Default exit code
------------- | -------------
error = nil   | 0 (success)
error != nil  | 1 (failure)
panic         | 1 (failure)

### License

[MIT](/LICENSE)

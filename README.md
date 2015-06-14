## Package closer ensures a clean exit for your Go app. 

The aim of this package is to provide an universal way to catch the event of application’s exit and perform some actions before it’s too late. `closer` doesn’t care about the way application tries to exit, i.e. was that a panic or just a signal from the OS, it calls the provided methods for cleanup and that’s the whole point.

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

The the usage examples: [example](/cmd/example/main.go), [example-error(/cmd/example-error/main.go) and [example-panic](/cmd/example-panic/main.go).

### Features

- [x] Ability to `Bind` custom functions to be called prior to app's termination;
- [x] Ability to signal the app termination event via the `Close` method (see the examples);
- [x] Ability to wrap functions with `Checked` in order to handle their panics;
- [x] Ability to override the default settings of this package with `Init`, like app exit codes and signals to catch.

### Table of exit codes

If you're using a simple catching method like in the [example](/cmd/example/main.go):
```go
defer closer.Close()
```
Then the exit code will always be a success (0) it that case. If you want to return the code which depends on the
function's return cause, you should use a helper like in the [example-error](/cmd/example-error/main.go) or the [example-panic](/cmd/example-panic/main.go):
```go
closer.Checked(run, true)
```

The errors and panics will be logged if the logging option was set true, also the exit code (for `os.Exit`) will be determined correctly:

Event  		  | Default exit code
------------- | -------------
error = nil   | 0 (success)
error != nil  | 1 (failure)
panic  		  | 1 (failure)


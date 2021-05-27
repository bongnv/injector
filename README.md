# injector

[![Go Reference](https://pkg.go.dev/badge/github.com/bongnv/injector.svg)](https://pkg.go.dev/github.com/bongnv/injector)
[![Build](https://github.com/bongnv/injector/workflows/CI/badge.svg)](https://github.com/bongnv/injector/actions?query=workflow%3ACI)
[![Go Report Card](https://goreportcard.com/badge/github.com/bongnv/injector)](https://goreportcard.com/report/github.com/bongnv/injector)
[![codecov](https://codecov.io/gh/bongnv/injector/branch/main/graph/badge.svg?token=RP3ua8huXh)](https://codecov.io/gh/bongnv/injector)

`injector` is a reflect-based dependency injection library for Go.

## Features

- Registering and injecting dependencies by names
- Injecting dependencies by types
- Registering dependencies by factory functions

## Getting started

1. Make sure Go is installed and add `injector` into your project by the following command:

```bash
go get github.com/bongnv/injector
```

2. Import it to your code:

```go
import "github.com/bongnv/injector"
```

3. Create a new instance of `Injector` and start registering, injecting dependencies.

```go
// ServiceAImpl is the example of an implementation.
type ServiceAImpl struct {}

// ServiceBImpl is another example of implementation that need to be injected.
type ServiceBImpl struct {
	// Here you can notice that ServiceBImpl requests a dependency with the type of *ServiceAImpl.
	ServiceA *ServiceAImpl `injector:"auto"`
}

func yourInitFunc() {
  i := injector.New()

  // add ServiceAImpl to the injector
  i.MustRegister(&ServiceAImpl{})

  // create an instance of ServiceBImpl and inject its dependencies
  b := &ServiceBImpl{}
  i.MustRegister(b)
}
```

## Usages

### Registering dependencies by factory functions

`loadAppConfig`, `newLogger` and `newServiceA` are three factory functions to create different components. `injector` allows loading dependencies as well as registering dependencies created by those functions.

```go
func loadAppConfig() *AppConfig {
  // load app config here
}

func newLogger(cfg *AppConfig) (Logger, error) {
  // initialize a logger
  return &loggerImpl{}
}

// ServiceA has Logger as a dependency
type ServiceA struct {
  Logger Logger `injector:"logger"`
}

func newServiceA() (*ServiceA, error) {
  // init your serviceA here
}

// init func
func initDependencies() {
  i := injector.New()
  i.MustRegisterNamed("config", loadAppConfig)
  i.MustRegisterNamed("logger", newLogger), 
  // serviceA will be created and registered, logger will also be injected
  i.MustRegister(newServiceA),
}
```

### Injecting dependencies by types

As `loggerImpl` satisfies the interface `Logger`, it will be injected into `ServiceA` automatically. If there are two dependencies that are eligible while injecting, an error will be returned. `auto` is the keyword to indicate the type-based injection.

`Unnamed` and `MustUnnamed` can be used to register dependencies if names is not used to identifying the dependency.

```go
// loggerImpl is an implementation that satisfies Logger interface.
type loggerImpl struct {}

// ServiceA has Logger as a dependency
type ServiceA struct {
  Logger Logger `injector:"auto"`
}

// init func
func initDependencies() {
  i := injector.New()
  i.MustRegisterNamed("logger", &loggerImpl{}), 
  // serviceA will be registered, logger will also be injected by Logger type
  i.MustRegister(&ServiceA{}),
}
```

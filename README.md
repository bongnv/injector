# injector

[![Go Reference](https://pkg.go.dev/badge/github.com/bongnv/injector.svg)](https://pkg.go.dev/github.com/bongnv/injector)
[![Build](https://github.com/bongnv/injector/workflows/Build/badge.svg)](https://github.com/bongnv/injector/actions?query=workflow%3ABuild)
[![Go Report Card](https://goreportcard.com/badge/github.com/bongnv/injector)](https://goreportcard.com/report/github.com/bongnv/injector)
[![codecov](https://codecov.io/gh/bongnv/injector/branch/main/graph/badge.svg?token=RP3ua8huXh)](https://codecov.io/gh/bongnv/injector)

`injector` is a reflect-based dependency injection library for Go.

## Features

- Registering and injecting dependencies by names
- Injecting dependencies by types
- Registering dependencies by factory functions

## Installation

```
go get github.com/bongnv/injector
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
  c := injector.New()
  c.MustRegister("config", loadAppConfig)
  c.MustRegister("logger", newLogger), 
  // serviceA will be created and registered, logger will also be injected
  c.MustRegister("serviceA", newServiceA),
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
  c := injector.New()
  c.MustRegister("logger", &loggerImpl{}), 
  // serviceA will be registered, logger will also be injected by Logger type
  c.MustRegister("serviceA", &ServiceA{}),
}
```

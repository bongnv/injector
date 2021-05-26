// Package injector provides a reflect-based injection solution where each dependency is
// identified by an unique name. A large application built with dependency
// injection in mind find difficulties in managing and injecting dependencies.
// This library attempts to take care of it by containing all dependencies in
// a central container and injecting requested dependencies automatically. Its use is
// simple that you use Register method to register a dependency. The library will
// search for tagged fields and try to inject requested dependencies.
//
// It works using Go's reflection package and is inherently limited in what it
// can do as opposed to a code-gen system with respect to private fields.
//
// The usage pattern for the library involves struct tags. It requires the tag
// format used by the various standard libraries, like json, xml etc. It
// involves tags in one of the form below:
//
//     `injector:"logger"`
//
// The above form is asking for a named dependency called "logger".
package injector

// New creates a new instance of Container.
func New() *Container {
	return &Container{
		dependencies: map[string]*dependency{},
	}
}

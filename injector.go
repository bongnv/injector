// Package injector provides a reflect-based injection solution where each dependency is
// identified by an unique name. A large application built with dependency
// injection in mind find difficulties in managing and injecting dependencies.
// This library attempts to take care of it by containing all dependencies in
// a central container and injecting requested dependencies automatically. Its use is
// simple that you use Component method to register a dependency. The library will
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

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	autoInjectionTag = "auto"
	unnamedPrefix    = "unnamed"
)

type dependency struct {
	value        interface{}
	reflectValue reflect.Value
	reflectType  reflect.Type
}

// Injector contains all dependencies. An injector can be created by New method.
type Injector struct {
	dependencies   map[string]*dependency
	unnamedCounter int
}

// NamedComponent registers new dependency with a name to the Injector. As name has to be unique,
// it returns an error if name is not unique. An error is also returned if the function is unable to inject dependencies.
// A factory function can be used:
//
// func newLogger() (Logger, error) {
//   // initialize a new logger
// }
//
// we then use c.NamedComponent("logger", newLogger) to register the logger dependency with that function.
// dependencies are also injected to the newly created struct from the factory function.
func (c *Injector) NamedComponent(name string, dep interface{}) error {
	if _, found := c.dependencies[name]; found {
		return fmt.Errorf("injector: %s is already registered", name)
	}

	if name == autoInjectionTag {
		return fmt.Errorf("injector: %s is revserved, please use a different name", autoInjectionTag)
	}

	var toAddDep *dependency
	depType := reflect.TypeOf(dep)
	if depType.Kind() == reflect.Func {
		createdDep, err := c.executeFunc(dep, depType)
		if err != nil {
			return err
		}

		toAddDep = createdDep
	} else {
		toAddDep = &dependency{
			value:        dep,
			reflectType:  depType,
			reflectValue: reflect.ValueOf(dep),
		}
	}

	if err := c.populate(toAddDep); err != nil {
		return err
	}

	c.dependencies[name] = toAddDep
	return nil
}

// MustNamedComponent is the similar to NamedComponent. Instead of returning an error,
// it panics if anything goes wrong.
func (c *Injector) MustNamedComponent(name string, dep interface{}) {
	if err := c.NamedComponent(name, dep); err != nil {
		panic(err)
	}
}

// Get loads a dependency from the Injector using name.
// It returns an error if the requested dependency couldn't be found.
func (c *Injector) Get(name string) (interface{}, error) {
	dep, found := c.dependencies[name]
	if !found {
		return nil, errors.New("injector: the requested dependency couldn't be found")
	}

	return dep.value, nil
}

// MustGet is the similar to Get. Instead of returning an error,
// it panics if anything goes wrong.
func (c *Injector) MustGet(name string) interface{} {
	dep, err := c.Get(name)
	if err != nil {
		panic(err)
	}

	return dep
}

// Component registers a new dependency without specifying the name.
// It's handy for injecting by types.
// One must be careful when injecting by types as it can cause conflicts easily.
func (c *Injector) Component(dep interface{}) error {
	for {
		newName := fmt.Sprintf("%s.%d", unnamedPrefix, c.unnamedCounter)
		if _, ok := c.dependencies[newName]; ok {
			c.unnamedCounter++
			continue
		}

		return c.NamedComponent(newName, dep)
	}
}

// MustComponent is similar to Component. Instead of returning an error, it will panic if there is any error.
func (c *Injector) MustComponent(dep interface{}) {
	if err := c.Component(dep); err != nil {
		panic(err)
	}
}

// Inject injects dependencies to a given object. It returns error if there is any.
// The object should be a pointer of struct, otherwise dependencies won't be injected.
func (c *Injector) Inject(object interface{}) error {
	dep := &dependency{
		value:        object,
		reflectType:  reflect.TypeOf(object),
		reflectValue: reflect.ValueOf(object),
	}

	return c.populate(dep)
}

func (c *Injector) populate(dep *dependency) error {
	if !isStructPtr(dep.reflectType) {
		if hasInjectTag(dep) {
			return fmt.Errorf("injector: %s is not injectable, a pointer is expected", dep.reflectType)
		}

		return nil
	}

	for i := 0; i < dep.reflectValue.Elem().NumField(); i++ {
		fieldValue := dep.reflectValue.Elem().Field(i)
		fieldType := fieldValue.Type()
		structField := dep.reflectType.Elem().Field(i)
		fieldTag := structField.Tag
		tagValue, ok := fieldTag.Lookup("injector")
		if !ok {
			continue
		}

		loadedDep, err := c.loadDepForTag(tagValue, fieldType)
		if err != nil {
			return err
		}

		if !loadedDep.reflectType.AssignableTo(fieldType) {
			return fmt.Errorf("injector: %s is not assignable from %s", fieldType, loadedDep.reflectType)
		}

		fieldValue.Set(loadedDep.reflectValue)
	}

	return nil
}

func (c *Injector) loadDepForTag(tag string, t reflect.Type) (*dependency, error) {
	if tag == autoInjectionTag {
		return c.findByType(t)
	}

	loadedDep, found := c.dependencies[tag]
	if !found {
		return nil, fmt.Errorf("injector: %s is not registered", tag)
	}

	return loadedDep, nil
}

func (c *Injector) executeFunc(fn interface{}, fnType reflect.Type) (*dependency, error) {
	if fnType.NumOut() > 2 || fnType.NumOut() < 1 {
		return nil, errors.New("injector: unsupported factory function")
	}

	if fnType.NumOut() == 2 && !implementsError(fnType.Out(1)) {
		return nil, errors.New("injector: 2nd output param must implement error")
	}

	fnVal := reflect.ValueOf(fn)
	inParams, err := c.generateInParams(fnType)
	if err != nil {
		return nil, err
	}

	out := fnVal.Call(inParams)
	if len(out) == 2 && !out[1].IsNil() {
		return nil, out[1].Interface().(error)
	}

	newDep := &dependency{
		value:        out[0].Interface(),
		reflectValue: out[0],
		reflectType:  out[0].Type(),
	}

	return newDep, nil
}

func (c *Injector) generateInParams(fnType reflect.Type) ([]reflect.Value, error) {
	params := make([]reflect.Value, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		param, err := c.findByType(fnType.In(i))
		if err != nil {
			return nil, err
		}

		params[i] = param.reflectValue
	}

	return params, nil
}

func (c *Injector) findByType(t reflect.Type) (*dependency, error) {
	var foundVal *dependency
	for _, v := range c.dependencies {
		if v.reflectType.AssignableTo(t) {
			if foundVal != nil {
				return nil, fmt.Errorf("injector: there is a conflict when finding the dependency for %s", t.String())
			}

			foundVal = v
		}
	}

	if foundVal == nil {
		return nil, fmt.Errorf("injector: couldn't find the dependency for %s", t.String())
	}

	return foundVal, nil
}

// New creates a new instance of Injector.
func New() *Injector {
	return &Injector{
		dependencies: map[string]*dependency{},
	}
}

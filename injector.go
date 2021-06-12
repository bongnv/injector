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

// Factory defines a factory that creates a new component.
type Factory interface {
	Create() (interface{}, error)
}

// New creates a new instance of Injector.
func New() *Injector {
	return &Injector{
		dependencies: map[string]*dependency{},
	}
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
func (c *Injector) NamedComponent(name string, dep interface{}) {
	c.validateNamne(name)

	toAddDep := &dependency{
		value:        dep,
		reflectType:  reflect.TypeOf(dep),
		reflectValue: reflect.ValueOf(dep),
	}

	if err := c.populate(toAddDep); err != nil {
		panic(err)
	}

	c.dependencies[name] = toAddDep
}

// NamedComponentFromFunc creates a new named component from a factory function
// and registers the created component to the injector.
func (c *Injector) NamedComponentFromFunc(name string, factoryFn interface{}) {
	c.validateNamne(name)

	fnType := reflect.TypeOf(factoryFn)
	if fnType.Kind() != reflect.Func {
		panic(errors.New("injector: a factory function is expected"))
	}

	createdDep, err := c.executeFunc(factoryFn, fnType)
	if err != nil {
		panic(err)
	}

	if err := c.populate(createdDep); err != nil {
		panic(err)
	}

	c.dependencies[name] = createdDep
}

// ComponentFromFunc creates a new component from a factory function.
// It's similar to NamedComponentFromFunc, instead a name will be generated for the component.
func (c *Injector) ComponentFromFunc(factoryFn interface{}) {
	c.NamedComponentFromFunc(c.nextGeneratedName(), factoryFn)
}

// ComponentFromFactory creates a new component by invoking the Create function in a given factory.
// Before creating the component, it will inject dependencies into the factory.
// After creating the component, it will inject dependencies to the component as well.
// It returns error if there is any.
//
// With ComponentFromFactory, the name will be generated for the generated component.
func (c *Injector) ComponentFromFactory(f Factory) {
	c.NamedComponentFromFactory(c.nextGeneratedName(), f)
}

// NamedComponentFromFactory creates a new component by invoking the Create function in a given factory.
// Before creating the component, it will inject dependencies into the factory.
// After creating the component, it will inject dependencies to the component as well.
func (c *Injector) NamedComponentFromFactory(name string, f Factory) {
	c.Inject(f)

	component, err := f.Create()
	if err != nil {
		panic(err)
	}

	c.NamedComponent(name, component)
}

// Get loads a dependency from the Injector using name.
func (c *Injector) Get(name string) interface{} {
	dep, found := c.dependencies[name]
	if !found {
		panic(errors.New("injector: the requested dependency couldn't be found"))
	}

	return dep.value
}

// Component registers a new dependency without specifying the name.
// It's handy for injecting by types.
// One must be careful when injecting by types as it can cause conflicts easily.
func (c *Injector) Component(dep interface{}) {
	c.NamedComponent(c.nextGeneratedName(), dep)
}

// Inject injects dependencies to a given object. It returns error if there is any.
// The object should be a pointer of struct, otherwise dependencies won't be injected.
func (c *Injector) Inject(object interface{}) {
	dep := &dependency{
		value:        object,
		reflectType:  reflect.TypeOf(object),
		reflectValue: reflect.ValueOf(object),
	}

	if err := c.populate(dep); err != nil {
		panic(err)
	}
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

func (c *Injector) nextGeneratedName() string {
	for {
		newName := fmt.Sprintf("%s.%d", unnamedPrefix, c.unnamedCounter)
		if _, ok := c.dependencies[newName]; !ok {
			return newName
		}
		c.unnamedCounter++
	}
}

func (c *Injector) validateNamne(name string) {
	if _, found := c.dependencies[name]; found {
		panic(fmt.Errorf("injector: %s is already registered", name))
	}

	if name == autoInjectionTag {
		panic(fmt.Errorf("injector: %s is revserved, please use a different name", autoInjectionTag))
	}
}

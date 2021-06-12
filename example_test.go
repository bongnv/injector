package injector_test

import (
	"fmt"

	"github.com/bongnv/injector"
)

// ServiceA is the example of an interface.
type ServiceA interface {
	Render()
}

// ServiceAImpl is the example of an implementation.
type ServiceAImpl struct {
	Data int
}

func (s *ServiceAImpl) Render() {
	fmt.Println("Data:", s.Data)
}

// ServiceBImpl is another example of implementation that need to be injected.
type ServiceBImpl struct {
	// Here you can notice that ServiceBImpl requests a dependency named "service-a".
	ServiceA ServiceA `injector:"service-a"`
}

func (s *ServiceBImpl) Render() {
	fmt.Println("Going to render ServiceA")
	s.ServiceA.Render()
	fmt.Println("ServiceA is rendered")
}

func Example() {
	// Typically an application will have exactly one instance of Injector, and
	// you will create it and use it in the initialization phase:
	c := injector.New()

	// Use NamedComponent to register an object into the Injector.
	c.NamedComponent("config", 10)

	// We can get a dependency by name and injectoror it manually.
	cfg := c.Get("config")

	// Initialize and register serviceA. cfg is injected manually.
	a := &ServiceAImpl{
		Data: cfg.(int),
	}

	// NamedComponent here is used to add another dependency to Injector.
	c.NamedComponent("service-a", a)

	// Initialize and register serviceB.
	b := &ServiceBImpl{}

	// NamedComponent will add b to Injector as well as inject a into b.ServiceA.
	c.NamedComponent("service-b", b)

	b.Render()
	// Output:
	// Going to render ServiceA
	// Data: 10
	// ServiceA is rendered
}

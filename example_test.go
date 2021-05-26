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

	// Use RegisterNamed to register an object into the Injector.
	errRegisterCfg := c.RegisterNamed("config", 10)
	if errRegisterCfg != nil {
		fmt.Println(errRegisterCfg)
	}

	// We can get a dependency by name and injectoror it manually.
	cfg, err := c.Get("config")
	if err != nil {
		fmt.Println(err)
	}

	// Initialize and register serviceA. cfg is injected manually.
	a := &ServiceAImpl{
		Data: cfg.(int),
	}

	// Register here is used to add another dependency to Injector.
	errRegisterA := c.RegisterNamed("service-a", a)
	if errRegisterA != nil {
		fmt.Println(errRegisterA)
	}

	// Initialize and register serviceB.
	b := &ServiceBImpl{}

	// Register will add b to Injector as well as inject a into b.ServiceA.
	errRegisterB := c.RegisterNamed("service-b", b)
	if errRegisterB != nil {
		fmt.Println(errRegisterB)
	}

	b.Render()
	// Output:
	// Going to render ServiceA
	// Data: 10
	// ServiceA is rendered
}

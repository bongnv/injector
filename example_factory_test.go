package injector_test

import "github.com/bongnv/injector"

// ServiceBFactory is a Factory implementation to create ServiceB
type ServiceBFactory struct{}

func (f ServiceBFactory) Create() (interface{}, error) {
	return &ServiceBImpl{}, nil
}

func ExampleFactory() {
	c := injector.New()
	c.NamedComponent("service-a", &ServiceAImpl{
		Data: 1,
	})

	c.NamedComponentFromFactory("service-b", &ServiceBFactory{})
	b := c.Get("service-b").(*ServiceBImpl)
	b.Render()
	// Output:
	// Going to render ServiceA
	// Data: 1
	// ServiceA is rendered
}

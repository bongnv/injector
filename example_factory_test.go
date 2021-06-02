package injector_test

import "github.com/bongnv/injector"

// ServiceBFactory is a Factory implementation to create ServiceB
type ServiceBFactory struct{}

func (f ServiceBFactory) Create() (interface{}, error) {
	return &ServiceBImpl{}, nil
}

func ExampleFactory() {
	c := injector.New()
	c.MustNamedComponent("service-a", &ServiceAImpl{
		Data: 1,
	})

	c.MustCreateNamedComponent("service-b", &ServiceBFactory{})
	b := c.MustGet("service-b").(*ServiceBImpl)
	b.Render()
	// Output:
	// Going to render ServiceA
	// Data: 1
	// ServiceA is rendered
}

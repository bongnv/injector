package injector

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type TypeA struct {
	Field int `injector:"mocked-int"`
}

type TypeB struct {
	Field *TypeA `injector:"type-a"`
}

type TypeC struct {
	Field TypeA `injector:"type-a"`
}

type TypeD struct {
	Field int `injector:"auto"`
}

func Test_NamedComponent(t *testing.T) {
	t.Run("happy-path", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		b := &TypeB{}
		c.NamedComponent("mocked-int", 10)
		c.NamedComponent("type-a", a)
		c.NamedComponent("type-b", b)
		require.EqualValues(t, 10, a.Field)
		require.EqualValues(t, a, b.Field)
	})

	t.Run("not-assignable", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		c.NamedComponent("mocked-int", "1000")
		require.PanicsWithError(t, "injector: int is not assignable from string", func() {
			c.NamedComponent("type-a", a)
		})
	})

	t.Run("not-assignable-pointer-expected", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		c.NamedComponent("mocked-int", 1000)
		c.NamedComponent("type-a", a)
		require.PanicsWithError(t, "injector: injector.TypeC is not injectable, a pointer is expected", func() {
			c.NamedComponent("type-c", TypeC{})
		})
	})

	t.Run("missing-dependency", func(t *testing.T) {
		c := New()
		b := &TypeB{}
		c.NamedComponent("mocked-int", "1000")
		require.PanicsWithError(t, "injector: type-a is not registered", func() {
			c.NamedComponent("type-b", b)
		})
	})

	t.Run("duplicate-registration", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		c.NamedComponent("type-a", 10)
		require.PanicsWithError(t, "injector: type-a is already registered", func() {
			c.NamedComponent("type-a", a)
		})
	})

	t.Run("duplicate-registration", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		c.NamedComponent("type-a", 10)
		require.PanicsWithError(t, "injector: type-a is already registered", func() {
			c.NamedComponent("type-a", a)
		})
	})
}

func Test_Get(t *testing.T) {
	t.Run("happy-path", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		c.NamedComponent("mocked-int", 10)
		c.NamedComponent("type-a", a)
		retrievedA := c.Get("type-a")
		require.IsType(t, &TypeA{}, retrievedA)
		require.EqualValues(t, 10, retrievedA.(*TypeA).Field)
	})

	t.Run("not-found", func(t *testing.T) {
		c := New()
		require.PanicsWithError(t, "injector: the requested dependency couldn't be found", func() {
			dep := c.Get("some-dep")
			require.Nil(t, dep)
		})
	})
}

func Test_ComponentFromFunc(t *testing.T) {
	t.Run("invalid-func", func(t *testing.T) {
		c := New()
		require.PanicsWithError(t, "injector: a factory function is expected", func() {
			c.ComponentFromFunc(10)
		})
	})

	t.Run("too-many-out-params", func(t *testing.T) {
		c := New()
		mockFunc := func() (int, int, error) {
			return 0, 0, nil
		}

		require.PanicsWithError(t, "injector: unsupported factory function", func() {
			c.ComponentFromFunc(mockFunc)
		})
	})

	t.Run("no-out-params", func(t *testing.T) {
		c := New()
		mockFunc := func() {}

		require.PanicsWithError(t, "injector: unsupported factory function", func() {
			c.ComponentFromFunc(mockFunc)
		})
	})

	t.Run("not-implements-error", func(t *testing.T) {
		c := New()
		mockFunc := func() (string, int) {
			return "", 0
		}

		require.PanicsWithError(t, "injector: 2nd output param must implement error", func() {
			c.ComponentFromFunc(mockFunc)
		})
	})

	t.Run("missing-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (int, error) {
			return 0, nil
		}

		require.PanicsWithError(t, "injector: couldn't find the dependency for string", func() {
			c.ComponentFromFunc(mockFunc)
		})
	})

	t.Run("conflict-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (int, error) {
			return 0, nil
		}

		c.NamedComponent("string-dep-1", "dep-1")
		c.NamedComponent("string-dep-2", "dep-2")
		require.PanicsWithError(t, "injector: there is a conflict when finding the dependency for string", func() {
			c.ComponentFromFunc(mockFunc)
		})
	})

	t.Run("error-with-factory-func", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (int, error) {
			return 0, errors.New("random error")
		}

		c.NamedComponent("string-dep-1", "dep-1")
		require.PanicsWithError(t, "random error", func() {
			c.ComponentFromFunc(mockFunc)
		})
	})

	t.Run("happy-path-with-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (*TypeA, error) {
			return &TypeA{}, nil
		}

		c.Component("dep-1")
		c.NamedComponent("mocked-int", 1)
		c.NamedComponentFromFunc("type-a", mockFunc)
		a := c.Get("type-a").(*TypeA)
		require.EqualValues(t, 1, a.Field)
	})

	t.Run("happy-path-without-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func() (int, error) {
			return 1, nil
		}

		c.NamedComponentFromFunc("int-dep", mockFunc)
		require.NotPanics(t, func() {
			require.EqualValues(t, 1, c.Get("int-dep"))
		}, "int-dep must be registered")
	})
}

func Test_NamedComponent_auto(t *testing.T) {
	c := New()
	c.NamedComponent("mocked-int", 10)
	d := &TypeD{}
	c.NamedComponent("type-d", d)
	require.Equal(t, 10, d.Field, "data should be injected by type")
}

func Test_NamedComponent_reserved_name(t *testing.T) {
	c := New()
	require.PanicsWithError(t, "injector: auto is revserved, please use a different name", func() {
		c.NamedComponent("auto", 10)
	})
}

func Test_Component(t *testing.T) {
	c := New()
	c.Component(10)
	require.Len(t, c.dependencies, 1)
	require.NotNil(t, c.dependencies["unnamed.0"])
	require.Equal(t, 0, c.unnamedCounter)
}

func Test_Component_taken(t *testing.T) {
	c := New()
	c.NamedComponent("unnamed.0", 10)
	c.Component(11)
	require.EqualValues(t, 11, c.Get("unnamed.1"))
}

func Test_Inject(t *testing.T) {
	c := New()
	c.Component(10)

	d := &TypeD{}
	c.Inject(d)
	require.Equal(t, 10, d.Field, "data should be injected")
}

type mockFactory struct {
	mockResult interface{}
	mockErr    error
}

func (m mockFactory) Create() (interface{}, error) {
	return m.mockResult, m.mockErr
}

type mockFactoryWithInjection struct {
	IntDep int `injector:"auto"`
}

func (m mockFactoryWithInjection) Create() (interface{}, error) {
	return "newObject", nil
}

func Test_NamedComponentFromFactory(t *testing.T) {
	c := New()
	c.NamedComponentFromFactory("component", &mockFactory{
		mockResult: "newObject",
	})
	require.Len(t, c.dependencies, 1)
	ret := c.Get("component")
	require.Equal(t, "newObject", ret)
}

func Test_ComponentFromFactory_error(t *testing.T) {
	c := New()
	require.PanicsWithError(t, "random error", func() {
		c.ComponentFromFactory(&mockFactory{
			mockErr: errors.New("random error"),
		})
	})
}

func Test_ComponentFromFactory_inject_failed(t *testing.T) {
	c := New()
	require.PanicsWithError(t, "injector: couldn't find the dependency for int", func() {
		c.ComponentFromFactory(&mockFactoryWithInjection{})
	})
}

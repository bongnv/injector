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

func Test_Register(t *testing.T) {
	t.Run("happy-path", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		b := &TypeB{}
		require.NoError(t, c.NamedComponent("mocked-int", 10))
		require.NoError(t, c.NamedComponent("type-a", a))
		require.NoError(t, c.NamedComponent("type-b", b))
		require.EqualValues(t, 10, a.Field)
		require.EqualValues(t, a, b.Field)
	})

	t.Run("not-assignable", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		require.NoError(t, c.NamedComponent("mocked-int", "1000"))
		err := c.NamedComponent("type-a", a)
		require.EqualError(t, err, "injector: int is not assignable from string")
	})

	t.Run("not-assignable-pointer-expected", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		require.NoError(t, c.NamedComponent("mocked-int", 1000))
		require.NoError(t, c.NamedComponent("type-a", a))
		err := c.NamedComponent("type-c", TypeC{})
		require.EqualError(t, err, "injector: injector.TypeC is not injectable, a pointer is expected")
	})

	t.Run("missing-dependency", func(t *testing.T) {
		c := New()
		b := &TypeB{}
		require.NoError(t, c.NamedComponent("mocked-int", "1000"))
		err := c.NamedComponent("type-b", b)
		require.EqualError(t, err, "injector: type-a is not registered")
	})

	t.Run("duplicate-registration", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		require.NoError(t, c.NamedComponent("type-a", 10))
		err := c.NamedComponent("type-a", a)
		require.EqualError(t, err, "injector: type-a is already registered")
	})

	t.Run("duplicate-registration", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		require.NoError(t, c.NamedComponent("type-a", 10))
		err := c.NamedComponent("type-a", a)
		require.EqualError(t, err, "injector: type-a is already registered")
	})
}

func Test_Get(t *testing.T) {
	t.Run("happy-path", func(t *testing.T) {
		c := New()
		a := &TypeA{}
		require.NoError(t, c.NamedComponent("mocked-int", 10))
		require.NoError(t, c.NamedComponent("type-a", a))
		retrievedA, err := c.Get("type-a")
		require.NoError(t, err)
		require.IsType(t, &TypeA{}, retrievedA)
		require.EqualValues(t, 10, retrievedA.(*TypeA).Field)
	})

	t.Run("not-found", func(t *testing.T) {
		c := New()
		dep, err := c.Get("some-dep")
		require.EqualError(t, err, "injector: the requested dependency couldn't be found")
		require.Nil(t, dep)
	})
}

func Test_MustRegister_panic(t *testing.T) {
	c := New()
	require.NoError(t, c.NamedComponent("mock-int", 10))
	require.Panics(t, func() {
		c.MustNamedComponent("mock-int", 20)
	})
}

func Test_MustGet_panic(t *testing.T) {
	c := New()
	require.Panics(t, func() {
		c.MustGet("some-dep")
	}, "it must panic as there is no request dep")
}

func Test_MustGet_no_panic(t *testing.T) {
	c := New()
	c.MustNamedComponent("mock-int", 10)
	require.NotPanics(t, func() {
		require.EqualValues(t, 10, c.MustGet("mock-int"))
	}, "it must panic as there is no request dep")
}

func Test_Register_factory_function(t *testing.T) {
	t.Run("too-many-out-params", func(t *testing.T) {
		c := New()
		mockFunc := func() (int, int, error) {
			return 0, 0, nil
		}

		err := c.NamedComponent("new-func", mockFunc)
		require.EqualError(t, err, "injector: unsupported factory function")
	})

	t.Run("no-out-params", func(t *testing.T) {
		c := New()
		mockFunc := func() {}

		err := c.NamedComponent("new-func", mockFunc)
		require.EqualError(t, err, "injector: unsupported factory function")
	})

	t.Run("not-implements-error", func(t *testing.T) {
		c := New()
		mockFunc := func() (string, int) {
			return "", 0
		}

		err := c.NamedComponent("new-func", mockFunc)
		require.EqualError(t, err, "injector: 2nd output param must implement error")
	})

	t.Run("missing-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (int, error) {
			return 0, nil
		}

		err := c.NamedComponent("int-dep", mockFunc)
		require.EqualError(t, err, "injector: couldn't find the dependency for string")
	})

	t.Run("conflict-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (int, error) {
			return 0, nil
		}

		c.MustNamedComponent("string-dep-1", "dep-1")
		c.MustNamedComponent("string-dep-2", "dep-2")
		err := c.NamedComponent("int-dep", mockFunc)
		require.EqualError(t, err, "injector: there is a conflict when finding the dependency for string")
	})

	t.Run("error-with-factory-func", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (int, error) {
			return 0, errors.New("random error")
		}

		c.MustNamedComponent("string-dep-1", "dep-1")
		err := c.NamedComponent("int-dep", mockFunc)
		require.EqualError(t, err, "random error")
	})

	t.Run("happy-path-with-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func(v string) (int, error) {
			return 1, nil
		}

		c.MustNamedComponent("string-dep-1", "dep-1")
		err := c.NamedComponent("int-dep", mockFunc)
		require.NoError(t, err)
		require.NotPanics(t, func() {
			require.EqualValues(t, 1, c.MustGet("int-dep"))
		}, "int-dep must be registered")
	})

	t.Run("happy-path-without-dependency", func(t *testing.T) {
		c := New()
		mockFunc := func() (int, error) {
			return 1, nil
		}

		err := c.NamedComponent("int-dep", mockFunc)
		require.NoError(t, err)
		require.NotPanics(t, func() {
			require.EqualValues(t, 1, c.MustGet("int-dep"))
		}, "int-dep must be registered")
	})
}

func Test_Register_auto(t *testing.T) {
	c := New()
	c.MustNamedComponent("mocked-int", 10)
	d := &TypeD{}
	err := c.NamedComponent("type-d", d)
	require.NoError(t, err)
	require.Equal(t, 10, d.Field, "data should be injected by type")
}

func Test_Register_reserved_name(t *testing.T) {
	c := New()
	err := c.NamedComponent("auto", 10)
	require.EqualError(t, err, "injector: auto is revserved, please use a different name")
}

func Test_Unnamed(t *testing.T) {
	c := New()
	err := c.Component(10)
	require.NoError(t, err, "New dependency shouldn't registered")
	require.Len(t, c.dependencies, 1)
	require.NotNil(t, c.dependencies["unnamed.0"])
	require.Equal(t, 0, c.unnamedCounter)
}

func Test_Unnamed_error(t *testing.T) {
	c := New()
	err := c.Component(&TypeA{})
	require.Error(t, err, "There should be error because of missing dependency")
	require.Equal(t, 0, c.unnamedCounter)
	require.NoError(t, c.Component(10))
	require.Len(t, c.dependencies, 1)
	require.Equal(t, 0, c.unnamedCounter)
	require.NotNil(t, c.dependencies["unnamed.0"], "Existing name should be reused")
	require.NoError(t, c.Component(11))
	require.NotNil(t, c.dependencies["unnamed.1"], "New name should be created")
}

func Test_MustUnnamed(t *testing.T) {
	c := New()
	require.Panics(t, func() {
		c.MustComponent(&TypeA{})
	}, "There must be panic because of missing dependency")
}

func Test_Inject(t *testing.T) {
	c := New()
	require.NoError(t, c.Component(10))

	d := &TypeD{}
	require.NoError(t, c.Inject(d), "there should be no error injecting dependencies")
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

func Test_CreateNamedComponent(t *testing.T) {
	c := New()
	require.NoError(t, c.CreateNamedComponent("component", &mockFactory{
		mockResult: "newObject",
	}))
	require.Len(t, c.dependencies, 1)
	ret, err := c.Get("component")
	require.NoError(t, err)
	require.Equal(t, "newObject", ret)
}

func Test_CreateComponent_error(t *testing.T) {
	c := New()
	err := c.CreateComponent(&mockFactory{
		mockErr: errors.New("random error"),
	})
	require.EqualError(t, err, "random error")
}

func Test_CreateNamedComponent_inject_failed(t *testing.T) {
	c := New()
	err := c.CreateComponent(&mockFactoryWithInjection{})
	require.EqualError(t, err, "injector: couldn't find the dependency for int")
}

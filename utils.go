package injector

import (
	"reflect"
)

var (
	reflectTypeOfError = reflect.TypeOf((*error)(nil)).Elem()
)

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func implementsError(t reflect.Type) bool {
	return t.Implements(reflectTypeOfError)
}

func hasInjectTag(dep *dependency) bool {
	if dep.reflectType.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < dep.reflectType.NumField(); i++ {
		structField := dep.reflectType.Field(i)
		if _, ok := structField.Tag.Lookup("injector"); ok {
			return true
		}
	}

	return false
}

package injector

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_hasInjectTag_no_tags(t *testing.T) {
	type structNoTag struct{}
	v := structNoTag{}
	require.False(t, hasInjectTag(&dependency{
		value:        v,
		reflectValue: reflect.ValueOf(v),
		reflectType:  reflect.TypeOf(v),
	}))
}

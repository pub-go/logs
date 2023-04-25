package arg_test

import (
	"fmt"
	"testing"

	"code.gopub.tech/logs/pkg/arg"
)

func TestJSON(t *testing.T) {
	type S struct {
		Age int
	}
	for _, tCase := range []struct {
		data interface{}
		want string
	}{
		{1, `1`},
		{"1", `"1"`},
		{S{20}, `{"Age":20}`},
		{nil, `null`},
	} {
		if got := fmt.Sprintf("%v", arg.JSON(tCase.data)); got != tCase.want {
			t.Errorf("got= %q want = %q", got, tCase.want)
		}
	}
	var m struct{ ID int } = struct{ ID int }{ID: 1}
	t.Logf("%%#v: %s", fmt.Sprintf("%#v", arg.JSON(m)))
	t.Logf("%%+v: %s", fmt.Sprintf("%+v", arg.JSON(m)))
	t.Logf(" %%v: %s", fmt.Sprintf("%v", arg.JSON(m)))
	t.Logf(" %%s: %s", fmt.Sprintf("%s", arg.JSON(m)))
}

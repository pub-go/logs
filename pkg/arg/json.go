package arg

import (
	"encoding/json"
	"fmt"
)

type Arg struct {
	data any
}

func JSON(data any) *Arg {
	return &Arg{data}
}

func (a *Arg) String() string {
	b, err := json.Marshal(a.data)
	if err != nil {
		return fmt.Sprintf("!(BADJSON|err=%+v|data=%#v)", err, a.data)
	}
	return string(b)
}

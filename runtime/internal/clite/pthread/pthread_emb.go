//go:build emb

package pthread

import c "github.com/goplus/llgo/runtime/internal/clite"

func Create(pthread *Thread, attr *Attr, routine RoutineFunc, arg c.Pointer) c.Int {}

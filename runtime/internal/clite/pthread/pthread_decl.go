package pthread

import c "github.com/goplus/llgo/runtime/internal/clite"

type aThread struct {
	Unused [8]byte
}

//llgo:type C
type RoutineFunc func(c.Pointer) c.Pointer

// Thread represents a POSIX thread.
type Thread = *aThread

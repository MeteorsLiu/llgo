//go:build emb

package sync

import (
	c "github.com/goplus/llgo/runtime/internal/clite"
	"github.com/goplus/llgo/runtime/internal/clite/time"
)

type MutexAttr struct {
}

func (a *MutexAttr) Init(attr *MutexAttr) c.Int { return 0 }

func (a *MutexAttr) Destroy() {}

func (a *MutexAttr) SetType(typ MutexType) c.Int { return 0 }

type Once struct {
}

var OnceInit Once

func (o *Once) Do(f func()) c.Int { return 0 }

type Mutex struct {
}

func (m *Mutex) Init(attr *MutexAttr) c.Int {
}

func (m *Mutex) Destroy() {
}

func (m *Mutex) TryLock() c.Int {
}

func (m *Mutex) Lock() {
}

func (m *Mutex) Unlock() {
}

// -----------------------------------------------------------------------------

// RWLockAttr is a read-write lock attribute object.
// pthread_rwlockattr_t
type RWLockAttr struct {
}

func (a *RWLockAttr) Init(attr *RWLockAttr) c.Int { return 0 }

func (a *RWLockAttr) Destroy() {}

func (a *RWLockAttr) SetPShared(pshared c.Int) c.Int { return 0 }

func (a *RWLockAttr) GetPShared(pshared *c.Int) c.Int { return 0 }

type RWLock struct {
}

func (rw *RWLock) Init(attr *RWLockAttr) c.Int { return 0 }

func (rw *RWLock) Destroy() {
}

func (rw *RWLock) RLock() {
}

func (rw *RWLock) TryRLock() c.Int { return 0 }

func (rw *RWLock) RUnlock() {
}

func (rw *RWLock) Lock() {
}

func (rw *RWLock) TryLock() c.Int { return 0 }

func (rw *RWLock) Unlock() {
}

// -----------------------------------------------------------------------------

// CondAttr is a condition variable attribute object.
// pthread_condattr_t
type CondAttr struct {
}

func (a *CondAttr) Init(attr *CondAttr) c.Int { return 0 }

func (a *CondAttr) Destroy() {}

// Cond is a condition variable.
// pthread_cond_t
type Cond struct {
}

func (c *Cond) Init(attr *CondAttr) c.Int {
	return 0
}

func (c *Cond) Destroy() {
}

func (c *Cond) Signal() c.Int {
	return 0
}

func (c *Cond) Broadcast() c.Int {
	return 0
}

func (c *Cond) Wait(m *Mutex) c.Int {
	return 0
}

func (c *Cond) TimedWait(m *Mutex, abstime *time.Timespec) c.Int {
	return 0
}

// -----------------------------------------------------------------------------

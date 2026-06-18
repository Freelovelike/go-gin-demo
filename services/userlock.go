package services

import "sync"

// userLocks serializes all authoritative writes for a single user so that
// concurrent requests (the client now uses several parallel HTTPRequest nodes)
// cannot interleave a check-then-mutate sequence — e.g. two plants both reading
// the same gold balance and both deducting it.
//
// Single-instance assumption: this is an in-process lock. If the server is ever
// scaled to multiple instances, replace this with a Redis-based distributed lock.
var userLocks sync.Map // map[uint]*sync.Mutex

// lockUser acquires the per-user lock and returns the unlock function.
//
//	defer lockUser(userID)()
func lockUser(userID uint) func() {
	actual, _ := userLocks.LoadOrStore(userID, &sync.Mutex{})
	mu := actual.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

package cache

import (
	"time"
)

type Details struct {
	added time.Time
}

//go:generate mockgen -source connectivityCaching.go -destination=mocks/connectivityCaching.go -package=mocks
type ConnectivityTester interface {
	// IsOkAndSet checks if the `u` value is contained, if it's not it checks it.
	// This operation should be thread safe, you can use it to modify the invalidatedCache state in the function.
	IsOkAndSet(testURL string, f func() bool) bool
	IsOk(testURL string) bool
	// Test if the operation can be completed with success. If so, invalidatedCache that.
	Test(testURL string) error
	Invalidate(url string)
}

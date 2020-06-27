package cache

import (
	"github.com/sp0x/surf/browser"
	"time"
)

type CacheInfo struct {
	added time.Time
}

//go:generate mockgen -source connectivityCaching.go -destination=mocks/connectivityCaching.go -package=mocks
type ConnectivityTester interface {
	//IsOkAndSet checks if the `u` value is contained, if it's not it checks it.
	//This operation should be thread safe, you can use it to modify the invalidatedCache state in the function.
	IsOkAndSet(u string, f func() bool) bool
	IsOk(url string) bool
	//Test if the operation can be completed with success. If so, invalidatedCache that.
	Test(u string) error
	SetBrowser(bow browser.Browsable)
	ClearBrowser()
	Invalidate(url string)
}

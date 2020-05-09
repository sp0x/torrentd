package storage

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"io/ioutil"
	"os"
)

var _ = Describe("Bolt storage", func() {
	It("Should be able to open a db", func() {
		db, err := GetBoltDb(tempfile())
		if err != nil {
			Fail(fmt.Sprintf("Couldn't open a db: %v", err))
			return
		}
		if db == nil {
			Fail(fmt.Sprintf("Nil db: %v", err))
			return
		}
		defer func() {
			_ = db.Close()
		}()
	})
	It("Should be able to store chat messages", func() {
		db, err := GetBoltDb(tempfile())
		if err != nil {
			Fail(fmt.Sprintf("Couldn't open a db: %v", err))
			return
		}
		if db == nil {
			Fail(fmt.Sprintf("Nil db: %v", err))
			return
		}
		bstore := BoltStorage{db}
		newchat := &Chat{
			"tester", "", 12,
		}
		err = bstore.StoreChat(newchat)
		if err != nil {
			Fail("Couldn't store chat")
			return
		}
		chat, err := bstore.GetChat(12)
		if err != nil {
			Fail("couldn't get chat from storage")
			return
		}
		if chat.ChatId != newchat.ChatId || chat.Username != newchat.Username || chat.InitialText != newchat.InitialText {
			Fail("couldn't properly restore chat from storage")
			return
		}

		defer func() {
			_ = db.Close()
		}()
	})
})

// tempfile returns a temporary file path.
func tempfile() string {
	f, err := ioutil.TempFile("", "bolt-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}

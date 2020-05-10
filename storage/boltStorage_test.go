package storage

import (
	"fmt"
	"github.com/boltdb/bolt"
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
	Context("with a database", func() {
		var db *bolt.DB
		bstore := &BoltStorage{}
		//Init db
		BeforeEach(func() {
			tempDb, err := GetBoltDb(tempfile())
			if err != nil {
				Fail(fmt.Sprintf("Couldn't open a db: %v", err))
				return
			}
			if tempDb == nil {
				Fail(fmt.Sprintf("Nil db: %v", err))
				return
			}
			db = tempDb
			bstore.Database = tempDb
		})
		//Teardown db
		AfterEach(func() {
			if db != nil {
				err := bstore.Truncate()
				if err != nil {
					Fail("couldn't teardown bolt storage")
				}
			}
		})
		It("Should be able to store chats", func() {
			newchat := &Chat{"tester", "", 12}
			err := bstore.StoreChat(newchat)
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
		})

		It("should return nil if a chat isn't found", func() {
			chat, err := bstore.GetChat(12)
			if err != nil {
				Fail(fmt.Sprintf("Error fetching non-existing chat: %v", err))
				return
			}
			if chat != nil {
				Fail("chat was not nil")
			}
		})

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

package indexing

func NewCursorOptions() *CursorOptions {
	return &CursorOptions{
		Limit: -1,
	}
}

func SingleItemCursor() *CursorOptions {
	return &CursorOptions{Limit: 1}
}

type CursorOptions struct {
	Limit   int
	Skip    int
	Reverse bool
}

type Cursor interface {
	First() ([]byte, []byte)
	Next() ([]byte, []byte)
	CanContinue(val []byte) bool
}

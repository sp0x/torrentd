package storage

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

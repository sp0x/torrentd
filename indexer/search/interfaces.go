package search

type Equatable interface {
	Equals(other interface{}) bool
}

type CanBeStale interface {
	SetState(new, updated bool)
}

type UUIDed interface {
	UUID() string
	SetUUID(string)
}

type IDed interface {
	Id() uint32
	SetId(uint32)
}

type WithUUIDAndId interface {
	UUIDed
	IDed
}
type Record interface {
	UUIDed
	IDed
	CanBeStale
}

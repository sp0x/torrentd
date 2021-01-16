package search

type Equatable interface {
	Equals(other interface{}) bool
}

type CanBeStale interface {
	SetState(new, updated bool)
	IsNew() bool
	IsUpdate() bool
}

type UUIDed interface {
	UUID() string
	SetUUID(string)
}

type IDed interface {
	GetID() uint32
	SetID(uint32)
}

type Record interface {
	UUIDed
	IDed
	CanBeStale
}

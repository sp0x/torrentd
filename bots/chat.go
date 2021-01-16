package bots

type Chat struct {
	Username    string
	InitialText string
	ChatID      int64
	UUIDValue   string
	RecordID    uint32
	isUpdate    bool
	isNew       bool
}

type ChatMessage struct {
	Text   string
	ChatID string
	Banner string
}

func (c *Chat) UUID() string {
	return c.UUIDValue
}

func (c *Chat) SetUUID(s string) {
	c.UUIDValue = s
}

func (c *Chat) GetID() uint32 {
	return c.RecordID
}

func (c *Chat) SetID(u uint32) {
	c.RecordID = u
}

func (c Chat) SetState(new, updated bool) {
	c.isNew = new
	c.isUpdate = updated
}

func (c *Chat) IsNew() bool {
	return c.isNew
}

func (c *Chat) IsUpdate() bool {
	return c.isUpdate
}

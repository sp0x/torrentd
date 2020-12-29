package bots

type Chat struct {
	Username    string
	InitialText string
	ChatId      int64
	UUIDValue   string
	RecordId    uint32
	isUpdate    bool
	isNew       bool
}

type ChatMessage struct {
	Text   string
	ChatId string
	Banner string
}

func (c *Chat) UUID() string {
	return c.UUIDValue
}

func (c *Chat) SetUUID(s string) {
	c.UUIDValue = s
}

func (c *Chat) Id() uint32 {
	return c.RecordId
}

func (c *Chat) SetId(u uint32) {
	c.RecordId = u
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

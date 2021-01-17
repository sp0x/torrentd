package indexer

type Details struct {
	ID       string
	Title    string
	Language string
	Link     string
}

func (i Details) GetID() string {
	return i.ID
}

func (i Details) GetTitle() string {
	return i.Title
}

func (i Details) GetLanguage() string {
	return i.Language
}

func (i Details) GetLink() string {
	return i.Link
}

func (r *Runner) Info() Info {
	return Details{
		ID:       r.definition.Site,
		Title:    r.definition.Name,
		Language: r.definition.Language,
		Link:     r.definition.Links[0],
	}
}

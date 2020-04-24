package indexer

type IndexerInfo struct {
	ID       string
	Title    string
	Language string
	Link     string
}

func (i IndexerInfo) GetId() string {
	return i.ID
}

func (i IndexerInfo) GetTitle() string {
	return i.Title
}

func (i IndexerInfo) GetLanguage() string {
	return i.Language
}

func (i IndexerInfo) GetLink() string {
	return i.Link
}

func (r *Runner) Info() Info {
	return IndexerInfo{
		ID:       r.definition.Site,
		Title:    r.definition.Name,
		Language: r.definition.Language,
		Link:     r.definition.Links[0],
	}
}

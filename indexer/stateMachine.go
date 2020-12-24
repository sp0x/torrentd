package indexer

func (r *Runner) initLogin() error {
	if r.definition.Login.Init.IsEmpty() {
		return nil
	}

	initUrl, err := r.getFullUrlInIndex(r.definition.Login.Init.Path)
	if err != nil {
		return err
	}

	return r.contentFetcher.FetchUrl(initUrl)
}

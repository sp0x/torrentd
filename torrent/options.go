package torrent

type GenericSearchOptions struct {
	PageCount            uint
	StartingPage         uint
	MaxRequestsPerSecond uint
	StopOnStaleTorrents  bool
}

type PaginationSearch struct {
	PageCount    uint
	StartingPage uint
}

type SearchRunOptions struct {
	MaxRequestsPerSecond uint
	StopOnStaleTorrents  bool
}

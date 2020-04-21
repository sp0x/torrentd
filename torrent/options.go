package torrent

type GenericSearchOptions struct {
	PageCount            uint
	StartingPage         uint
	MaxRequestsPerSecond uint
	StopOnStaleTorrents  bool
}

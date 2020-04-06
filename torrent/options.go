package torrent

type FetchOptions struct {
	PageCount            uint
	StartingPage         uint
	MaxRequestsPerSecond uint
	StopOnStaleTorrents  bool
}

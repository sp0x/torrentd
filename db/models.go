package db

//type Torrent struct {
//	gorm.Model
//	Name         string
//	TorrentId    string
//	AddedOn      int64 // *time.Time
//	Link         string
//	Fingerprint  string
//	AuthorName   string
//	AuthorID     string
//	CategoryName string
//	CategoryID   string
//	Size         uint64
//	Seeders      int
//	Leachers     int
//	Downloaded   int
//	DownloadLink string
//	IsMagnet     bool
//	Announce     string
//	Publisher    string
//	AltName      string
//}

type TorrentCategory struct {
	CategoryID   string
	CategoryName string
}

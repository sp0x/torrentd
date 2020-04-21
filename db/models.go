package db

//type Torrent struct {
//	gorm.Model
//	Name         string
//	TorrentId    string
//	AddedOn      int64 // *time.Time
//	Link         string
//	Fingerprint  string
//	AuthorName   string
//	AuthorId     string
//	CategoryName string
//	CategoryId   string
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
	CategoryId   string
	CategoryName string
}

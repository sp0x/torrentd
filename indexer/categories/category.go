package categories

import "fmt"

type Category struct {
	ID   int
	Name string
}

func (c Category) String() string {
	return fmt.Sprintf("%s[%d]", c.Name, c.ID)
}

const (
	CustomCategoryOffset = 100000
)

var Uncategorized = &Category{ID: 0, Name: "uncategorized"}

// Categories from the Newznab spec
// https://github.com/nZEDb/nZEDb/blob/0.x/docs/newznab_api_specification.txt#L627
var (
	Subtitle                  = Category{-1, "Subtitle"}
	Rental                    = Category{-2, "Rental"}
	CategoryOther             = Category{0, "Other"}
	CategoryOtherMisc         = Category{10, "Other/Misc"}
	CategoryOtherHashed       = Category{20, "Other/Hashed"}
	CategoryConsole           = Category{1000, "Console"}
	CategoryConsoleNDS        = Category{1010, "Console/NDS"}
	CategoryConsolePSP        = Category{1020, "Console/PSP"}
	CategoryConsoleWii        = Category{1030, "Console/Wii"}
	CategoryConsoleXBOX       = Category{1040, "Console/Xbox"}
	CategoryConsoleXBOX360    = Category{1050, "Console/Xbox360"}
	CategoryConsoleWiiwareVC  = Category{1060, "Console/Wiiware/V"}
	CategoryConsoleXBOX360DLC = Category{1070, "Console/Xbox360"}
	CategoryConsolePS3        = Category{1080, "Console/PS3"}
	CategoryConsoleOther      = Category{1999, "Console/Other"}
	CategoryConsole3DS        = Category{1110, "Console/3DS"}
	CategoryConsolePSVita     = Category{1120, "Console/PS Vita"}
	CategoryConsoleWiiU       = Category{1130, "Console/WiiU"}
	CategoryConsoleXBOXOne    = Category{1140, "Console/XboxOne"}
	CategoryConsolePS4        = Category{1180, "Console/PS4"}
	CategoryMovies            = Category{2000, "Movies"}
	CategoryMoviesForeign     = Category{2010, "Movies/Foreign"}
	CategoryMoviesOther       = Category{2020, "Movies/Other"}
	CategoryMoviesSD          = Category{2030, "Movies/SD"}
	CategoryMoviesHD          = Category{2040, "Movies/HD"}
	CategoryMoviesUHD         = Category{2045, "Movies/UHD"}
	CategoryMovies3D          = Category{2050, "Movies/3D"}
	CategoryMoviesBluRay      = Category{2060, "Movies/BluRay"}
	CategoryMoviesDVD         = Category{2070, "Movies/DVD"}
	CategoryMoviesWEBDL       = Category{2080, "Movies/WEBDL"}
	CategoryMoviesTrailers    = Category{2090, "Movies/Trailers"}
	CategoryAudio             = Category{3000, "Audio"}
	CategoryAudioMP3          = Category{3010, "Audio/MP3"}
	CategoryAudioVideo        = Category{3020, "Audio/Video"}
	CategoryAudioAudiobook    = Category{3030, "Audio/Audiobook"}
	CategoryAudioLossless     = Category{3040, "Audio/Lossless"}
	CategoryAudioOther        = Category{3999, "Audio/Other"}
	CategoryAudioForeign      = Category{3060, "Audio/Foreign"}
	CategoryPC                = Category{4000, "PC"}
	CategoryPC0day            = Category{4010, "PC/0day"}
	CategoryPCISO             = Category{4020, "PC/ISO"}
	CategoryPCMac             = Category{4030, "PC/Mac"}
	CategoryPCPhoneOther      = Category{4040, "PC/Phone-Other"}
	CategoryPCGames           = Category{4050, "PC/Games"}
	CategoryPCPhoneIOS        = Category{4060, "PC/Phone-IOS"}
	CategoryPCPhoneAndroid    = Category{4070, "PC/Phone-Android"}
	CategoryTV                = Category{5000, "TV"}
	CategoryTVWEBDL           = Category{5010, "TV/WEB-DL"}
	CategoryTVFOREIGN         = Category{5020, "TV/Foreign"}
	CategoryTVDVD             = Category{5025, "TV/DVD"}
	CategoryTVSD              = Category{5030, "TV/SD"}
	CategoryTVHD              = Category{5040, "TV/HD"}
	CategoryTVUHD             = Category{5045, "TV/UHD"}
	CategoryTVOther           = Category{5999, "TV/Other"}
	CategoryTVSport           = Category{5060, "TV/Sport"}
	CategoryTVAnime           = Category{5070, "TV/Anime"}
	CategoryTVDocumentary     = Category{5080, "TV/Documentary"}
	CategoryXXX               = Category{6000, "XXX"}
	CategoryXXXDVD            = Category{6010, "XXX/DVD"}
	CategoryXXXWMV            = Category{6020, "XXX/WMV"}
	CategoryXXXXviD           = Category{6030, "XXX/XviD"}
	CategoryXXXx264           = Category{6040, "XXX/x264"}
	CategoryXXXOther          = Category{6999, "XXX/Other"}
	CategoryXXXImageset       = Category{6060, "XXX/Imageset"}
	CategoryXXXPacks          = Category{6070, "XXX/Packs"}
	CategoryBooks             = Category{7000, "Books"}
	CategoryBooksMagazines    = Category{7010, "Books/Magazines"}
	CategoryBooksEbook        = Category{7020, "Books/Ebook"}
	CategoryBooksComics       = Category{7030, "Books/Comics"}
	CategoryBooksTechnical    = Category{7040, "Books/Technical"}
	CategoryBooksForeign      = Category{7060, "Books/Foreign"}
	CategoryBooksUnknown      = Category{7999, "Books/Unknown"}
)

var AllCategories = CreateCategorySet([]Category{
	Subtitle,
	Rental,
	CategoryOther,
	CategoryOtherMisc,
	CategoryOtherHashed,
	CategoryConsole,
	CategoryConsoleNDS,
	CategoryConsolePSP,
	CategoryConsoleWii,
	CategoryConsoleXBOX,
	CategoryConsoleXBOX360,
	CategoryConsoleWiiwareVC,
	CategoryConsoleXBOX360DLC,
	CategoryConsolePS3,
	CategoryConsoleOther,
	CategoryConsole3DS,
	CategoryConsolePSVita,
	CategoryConsoleWiiU,
	CategoryConsoleXBOXOne,
	CategoryConsolePS4,
	CategoryMovies,
	CategoryMoviesForeign,
	CategoryMoviesOther,
	CategoryMoviesSD,
	CategoryMoviesHD,
	CategoryMoviesUHD,
	CategoryMovies3D,
	CategoryMoviesBluRay,
	CategoryMoviesDVD,
	CategoryMoviesWEBDL,
	CategoryMoviesTrailers,
	CategoryAudio,
	CategoryAudioMP3,
	CategoryAudioVideo,
	CategoryAudioAudiobook,
	CategoryAudioLossless,
	CategoryAudioOther,
	CategoryAudioForeign,
	CategoryPC,
	CategoryPC0day,
	CategoryPCISO,
	CategoryPCMac,
	CategoryPCPhoneOther,
	CategoryPCGames,
	CategoryPCPhoneIOS,
	CategoryPCPhoneAndroid,
	CategoryTV,
	CategoryTVWEBDL,
	CategoryTVFOREIGN,
	CategoryTVDVD,
	CategoryTVSD,
	CategoryTVHD,
	CategoryTVUHD,
	CategoryTVOther,
	CategoryTVSport,
	CategoryTVAnime,
	CategoryTVDocumentary,
	CategoryXXX,
	CategoryXXXDVD,
	CategoryXXXWMV,
	CategoryXXXXviD,
	CategoryXXXx264,
	CategoryXXXOther,
	CategoryXXXImageset,
	CategoryXXXPacks,
	CategoryBooks,
	CategoryBooksMagazines,
	CategoryBooksEbook,
	CategoryBooksComics,
	CategoryBooksTechnical,
	CategoryBooksForeign,
	CategoryBooksUnknown,
})

type Categories map[int]*Category

func CreateCategorySet(cats []Category) Categories {
	cs := Categories{}
	for _, c := range cats {
		cx := c
		cs[c.ID] = &cx
	}
	return cs
}

func ParentCategory(c *Category) Category {
	switch {
	case c.ID < 1000:
		return CategoryOther
	case c.ID < 2000:
		return CategoryConsole
	case c.ID < 3000:
		return CategoryMovies
	case c.ID < 4000:
		return CategoryAudio
	case c.ID < 5000:
		return CategoryPC
	case c.ID < 6000:
		return CategoryTV
	case c.ID < 7000:
		return CategoryXXX
	case c.ID < 8000:
		return CategoryBooks
	}
	return CategoryOther
}

func (slice Categories) Items() []*Category {
	v := make([]*Category, 0, len(slice))
	for _, c := range slice {
		if c == nil {
			continue
		}
		v = append(v, c)
	}
	return v
}

func (slice Categories) ContainsCat(cat *Category) bool {
	_, ok := slice[cat.ID]
	return ok
}

func (slice Categories) Subset(ids ...int) Categories {
	cats := Categories{}

	for _, cat := range AllCategories {
		cat := cat
		for _, id := range ids {
			if cat.ID == id {
				cats[cat.ID] = cat
			}
		}
	}

	return cats
}

func (slice Categories) Len() int {
	return len(slice)
}

func (slice Categories) Less(i, j int) bool {
	return slice[i].ID < slice[j].ID
}

func (slice Categories) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

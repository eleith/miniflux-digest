package models

import (
	miniflux "miniflux.app/v2/client"
	"time"
)

type FeedIcon struct {
	FeedID int64
	Data   string
}

type HTMLTemplateData struct {
	Category      *miniflux.Category
	Entries       *miniflux.Entries
	GeneratedDate time.Time
	FeedIcons     []*FeedIcon
	EntryGroups   []*EntryGroup
}

type EntryGroup struct {
	Title   string
	Entries []*miniflux.Entry
}

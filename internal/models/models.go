package models

import (
	miniflux "miniflux.app/v2/client"
	"time"
)

type FeedIcon struct {
	FeedID int64
	Data   string
}

type OverviewTemplateData struct {
	Category      *miniflux.Category
	Entries       *miniflux.Entries
	GeneratedDate time.Time
	FeedIcons     []*FeedIcon
	EntryGroups   []*EntryGroup
	OverviewSummary       string
	MinifluxHost  string
}

type EntryGroup struct {
	Title   string
	Summary string
	Entries []*miniflux.Entry
}

type GroupTemplateData struct {
	Title         string
	Summary       string
	Entries       []*miniflux.Entry
	GeneratedDate time.Time
	FeedIcons     []*FeedIcon
	MinifluxHost  string
}

type RawCategoryData struct {
	Category *miniflux.Category
	Entries  *miniflux.Entries
	Feeds    []*miniflux.Feed
	Icons    map[int64]*FeedIcon
	Err      error
}

type GroupedEntriesTemplateData struct {
	EntryGroup    *EntryGroup
	GeneratedDate time.Time
	FeedIcons     []*FeedIcon
	MinifluxHost  string
}

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
	PrimaryGroups []*PrimaryGroupDigestData
	OverviewSummary       string
	MinifluxHost  string
}

type EntryGroup struct {
	Title   string
	Summary string
	Entries []*miniflux.Entry
	Slug    string
}

type GroupTemplateData struct {
	Title         string
	Summary       string
	Entries       []*miniflux.Entry
	GeneratedDate time.Time
	FeedIcons     []*FeedIcon
	MinifluxHost  string
}



type GroupedEntriesTemplateData struct {
	EntryGroup    *EntryGroup
	GeneratedDate time.Time
	FeedIcons     []*FeedIcon
	MinifluxHost  string
}

type PrimaryGroupDigestData struct {
	Title     string
	Slug      string
	SubGroups []*EntryGroup
	Summary   string
}

type GroupedDigestPageData struct {
	PrimaryGroup  *PrimaryGroupDigestData
	FeedIcons     []*FeedIcon
	MinifluxHost  string
	GeneratedDate time.Time
}

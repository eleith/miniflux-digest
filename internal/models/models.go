package models

import (
	"time"
)

type Entry struct {
	ID         int64
	Title      string
	URL        string
	Content    string
	FeedID     int64
	FeedTitle  string
	GroupID    int64
	GroupTitle string
	CommentsURL string
	Date       time.Time
}

type FeedIcon struct {
	FeedID int64
	Data   string
}

type OverviewTemplateData struct {
	Entries       []*Entry
	GeneratedDate time.Time
	FeedIcons     []*FeedIcon
	EntryGroups   []*EntryGroup
	PrimaryGroups []*PrimaryGroupDigestData
	OverviewSummary       string
	TotalEntries          int
	TotalFeeds            int
	MinifluxHost  string
}

type EntryGroup struct {
	Title        string
	Summary      string
	Entries      []*Entry
	Slug         string
	TotalEntries int
	TotalFeeds   int
}

type GroupTemplateData struct {
	Title         string
	Summary       string
	Entries       []*Entry
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
	ID           int64
	Title        string
	Slug         string
	SubGroups    []*EntryGroup
	Summary      string
	TotalEntries int
	TotalFeeds   int
}

type GroupedDigestPageData struct {
	PrimaryGroup  *PrimaryGroupDigestData
	FeedIcons     []*FeedIcon
	MinifluxHost  string
	GeneratedDate time.Time
	TotalEntries  int
	TotalFeeds    int
}

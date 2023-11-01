package main

type Folder struct {
	ID          string
	Name        string
	Description string
	Tags        string
	Rank        int
	RankMode    string
	Articles    []Article

	CreatedTime int64
}

type Category struct {
	ID          string
	FolderID    string
	Name        string
	Rank        int
	Description string

	CreatedTime int64
	UpdateTime  int64
}

type Article struct {
	ID         string
	Title      string
	Content    string
	Summary    string
	Count      int
	FolderID   string
	CategoryID string
	Rank       int

	CreateTime int64
	UpdateTime int64
}

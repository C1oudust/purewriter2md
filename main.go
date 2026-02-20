package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/glebarez/go-sqlite"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var needMeta = false
var db *sql.DB

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Println("Usage: pw2md.exe <purewriter .db filepath>")
		os.Exit(1)
	}

	filePath := args[1]
	filename := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	if filename == "" {
		filename = "purewriter_export"
	}

	var err error
	db, err = sql.Open("sqlite", filePath)
	if err != nil {
		log.Fatal(err)
	}
	var flag string

	log.Printf("Output md file time metadata?(y/n): ")
	_, err = fmt.Scanln(&flag)
	if err != nil {
		log.Fatal("input err:", err)
		return
	}

	if flag == "Y" || flag == "y" {
		needMeta = true
	}

	log.Println("reading purewriter database...")
	defer db.Close()

	var folderList []Folder
	query := `SELECT f.id, f.name, f.createdTime, COALESCE(f.description, '') as description, COALESCE(f.tags, '') as tags, f.rank, COALESCE(f.rankMode, '') as rankMode FROM Folder f WHERE f.id!=? AND COALESCE(f.deleted, 0)=0`
	rows, err := db.Query(query, "PW_Trash")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var folder Folder
		if err = rows.Scan(&folder.ID, &folder.Name, &folder.CreatedTime, &folder.Description, &folder.Tags, &folder.Rank, &folder.RankMode); err != nil {
			log.Fatal(err)
		}

		folder.Articles, err = LoadArticles(folder.ID)
		if err != nil {
			log.Fatal(err)
		}

		folder.Categories, err = LoadCategories(folder.ID)
		if err != nil {
			log.Fatal(err)
		}

		folderList = append(folderList, folder)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	CreateFolder(filename, folderList)

	log.Println("done. press any key to exit.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func LoadArticles(folderID string) ([]Article, error) {
	query := `SELECT a.id, COALESCE(a.title, '') as title, a.content, COALESCE(a.summary, '') as summary, COALESCE(a.count, 0) as count, a.extension, a.folderId, COALESCE(a.categoryId, '') as categoryId, a.rank, a.updateTime, a.createTime FROM Article a WHERE a.folderId=? AND COALESCE(a.deleted, 0)=0 ORDER BY a.rank ASC, a.createTime ASC`
	rows, err := db.Query(query, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var article Article
		if err = rows.Scan(&article.ID, &article.Title, &article.Content, &article.Summary, &article.Count, &article.Extension, &article.FolderID, &article.CategoryID, &article.Rank, &article.UpdateTime, &article.CreateTime); err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return articles, nil
}

func LoadCategories(folderID string) ([]Category, error) {
	query := `SELECT c.id, COALESCE(c.name, '') as name, c.folderID, COALESCE(c.description, '') as description, c.rank, c.updateTime, c.createdTime FROM Category c WHERE c.folderId=? AND COALESCE(c.deleted, 0)=0 ORDER BY c.rank ASC`
	rows, err := db.Query(query, folderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err = rows.Scan(&category.ID, &category.Name, &category.FolderID, &category.Description, &category.Rank, &category.UpdateTime, &category.CreatedTime); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

func CreateFolder(rootPath string, folderList []Folder) {
	_, err := os.Stat(rootPath)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(rootPath, 0755)
	}
	for _, folder := range folderList {
		folderPath := strings.Trim(folder.Name, " ")
		outPath := path.Join(rootPath, folderPath)

		_ = os.MkdirAll(outPath, 0755)
		err = CreateFolderMeta(folder, outPath)
		if err != nil {
			log.Println("create meta.json failed:", err)
		}
		log.Println("parse book:", folder.Name)

		if folder.RankMode == "RANK" || len(folder.Categories) > 0 {
			CreateCategory(folder, outPath)
			continue
		}
		CreateArticles(folder.Articles, outPath)
	}
}

func CreateFolderMeta(folder Folder, outPath string) error {
	file, _ := os.Create(path.Join(outPath, "meta.json"))
	defer file.Close()
	data, _ := json.MarshalIndent(map[string]interface{}{
		"id":          folder.ID,
		"name":        folder.Name,
		"createdTime": ParseTime(folder.CreatedTime, ""),
		"description": folder.Description,
		"tags":        folder.Tags,
	}, "", "  ")
	_, err = file.Write(data)
	return err
}

func CreateArticleMeta(article Article) (meta string) {
	createTime := ParseTime(article.CreateTime, "")
	updateTime := ParseTime(article.UpdateTime, "")
	if needMeta {
		meta = fmt.Sprintf(`---
create: %s
update: %s
---

`, createTime, updateTime)
	}
	return
}

func CreateArticles(articles []Article, outPath string) {
	for _, article := range articles {
		filename := article.Title
		if filename == "" {
			filename = "Untitled-" + ParseTime(article.CreateTime, "2006_01_02_15_04_05")
		}
		regex := regexp.MustCompile(`[\\/:*?"<>|]`)
		// replace invalid characters to 'x'
		filename = regex.ReplaceAllString(filename, "x")
		filePath := filepath.Join(outPath, filename+".md")
		file, err := os.Create(filePath)
		if err != nil {
			log.Println("create md failed:", err)
			continue
		}
		defer file.Close()
		content := article.Content
		if article.Extension == "txt" {
			// markdown line wrapping
			content = strings.ReplaceAll(article.Content, "\n", "\n\n")
		}

		_, err = file.WriteString(CreateArticleMeta(article) + content)
		if err != nil {
			log.Println("write md failed:", err)
		}
	}
}

func CreateCategory(folder Folder, outPath string) {
	if len(folder.Categories) == 0 {
		CreateArticles(folder.Articles, outPath)
		return
	}

	categories := append([]Category(nil), folder.Categories...)
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Rank < categories[j].Rank
	})

	for _, category := range categories {
		folderPath := strings.Trim(category.Name, " ")
		categoryPath := path.Join(outPath, folderPath)
		_ = os.MkdirAll(categoryPath, 0755)

		var curArticleList []Article
		for _, a := range folder.Articles {
			if math.Abs(float64(a.Rank-category.Rank)) < 9999 {
				curArticleList = append(curArticleList, a)
			}
		}
		sort.Slice(curArticleList, func(i, j int) bool {
			return curArticleList[i].Rank < curArticleList[j].Rank
		})
		CreateArticles(curArticleList, categoryPath)
	}
}

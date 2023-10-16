package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/glebarez/go-sqlite"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Folder struct {
	ID          string
	Name        string
	CreatedTime int64
	Description string
	Tags        string
	Articles    []Article
}

type Article struct {
	ID         string
	Title      string
	Content    string
	Summary    string
	Count      int
	FolderID   string
	UpdateTime int64
	CreateTime int64
}

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Println("Usage: pw2md.exe <purewriter .db filepath>")
		os.Exit(1)
	}

	filePath := args[1]
	filename := strings.Split(filepath.Base(filePath), ".")[0]
	db, err := sql.Open("sqlite", filePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("reading purewriter database...")
	defer db.Close()

	var folderList []Folder
	query := `SELECT f.id, f.name, f.createdTime, COALESCE(f.description, '') as description,COALESCE(f.tags, '') as tags FROM Folder f WHERE NOT f.id=?`
	rows, err := db.Query(query, "PW_Trash")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var folder Folder
		if err = rows.Scan(&folder.ID, &folder.Name, &folder.CreatedTime, &folder.Description, &folder.Tags); err != nil {
			log.Fatal(err)
		}
		query = `SELECT a.id, COALESCE(a.title, '') as title, a.content, COALESCE(a.summary, '') as summary, COALESCE(a.count, 0) as  count, a.folderId, a.updateTime, a.createTime FROM  Article a WHERE a.folderId=?`
		as, err := db.Query(query, folder.ID)
		if err != nil {
			log.Fatal(err)
		}
		for as.Next() {
			var article Article
			if err = as.Scan(&article.ID, &article.Title, &article.Content, &article.Summary, &article.Count, &article.FolderID, &article.UpdateTime, &article.CreateTime); err != nil {
				log.Fatal(err)
			}
			folder.Articles = append(folder.Articles, article)
		}
		folderList = append(folderList, folder)
	}
	CreateFolder(filename, folderList)

	log.Println("done. press any key to exit.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func CreateFolder(folderName string, folderList []Folder) {
	_, err := os.Stat(folderName)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(folderName, 0755)
	}
	for _, folder := range folderList {

		folderPath := strings.Trim(folder.Name, " ")
		outPath := path.Join(folderName, folderPath)

		_ = os.MkdirAll(outPath, 0755)
		err := CreateFolderMeta(folder, outPath)
		if err != nil {
			log.Println("create meta.json failed:", err)
		}
		log.Println("parse folder:", folder.Name)
		for _, article := range folder.Articles {
			filename := article.Title
			if filename == "" {
				filename = "Untitled-" + time.Unix(article.CreateTime/1000, 0).Format("2006_01_02_15_04_05")
			}

			filePath := filepath.Join(outPath, filename+".md")
			file, err := os.Create(filePath)
			if err != nil {
				log.Println("create md failed:", err)
				continue
			}
			defer file.Close()
			// markdown line wrapping
			content := strings.ReplaceAll(article.Content, "\n", "\n\n")
			_, err = file.WriteString(CreateArticleMeta(article) + content)
			if err != nil {
				log.Println("write md failed:", err)
			}
		}
	}
}

func CreateFolderMeta(folder Folder, outPath string) error {
	file, _ := os.Create(path.Join(outPath, "meta.json"))
	defer file.Close()
	data, _ := json.MarshalIndent(map[string]interface{}{
		"id":          folder.ID,
		"name":        folder.Name,
		"createdTime": folder.CreatedTime,
		"description": folder.Description,
		"tags":        folder.Tags,
	}, "", "  ")
	_, err := file.Write(data)
	return err
}

func CreateArticleMeta(article Article) (meta string) {
	createTime := time.Unix(article.CreateTime/1000, 0).Format("2006-01-02 15:04:05")
	updateTime := time.Unix(article.UpdateTime/1000, 0).Format("2006-01-02 15:04:05")

	meta = fmt.Sprintf(`---
create: %s
update: %s
---

`, createTime, updateTime)
	return
}

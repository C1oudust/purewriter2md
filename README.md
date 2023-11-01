# 📃purewriter2md

[中文文档](./README_cn.md)

purewriter2md is a tool to export PureWriter database to Markdown file

## Usage

### get PureWriter database

- get your PureWriter backup file in local or cloud storage
- change file name from `*.pwb` to `*.rar`
- unzip `*.rar` and there is database `*.db` file

### export markdown from database

- drag your `*.db` file to `pw2md.exe` or run
    ```bash
    pw2md.exe /path/to/*.db 
    ```

## Other

- Each book will be exported as a folder, and the specific information of the book will be exported as a separate
  meta.json, including the book's description, tags, etc.
- I use this tool to export Markdown in order to quickly migrate to Obsidian.

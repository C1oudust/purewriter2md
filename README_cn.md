# 📃purewriter2md

[English](./README.md)

一个将纯纯写作数据导出为Markdown文件的工具

## 使用

### 获取纯纯写作数据库

- 从本地或云备份中获取到备份文件，通常以`pwd`结尾
- 将文件后缀改为`.rar`
- 解压，可以看到和压缩包名同名的`.db`文件，这就是纯纯写作的数据库

### 从数据库导出为markdown

- 将 `*.db` 文件拖拽到 `pw2md.exe` 上或者命令行运行
    ```bash
    pw2md.exe db文件的绝对路径
    ```

## 其他

- 每本书会导出为一个文件夹，书本的具体信息会单独导出为一个meta.json，包括书本的描述、标签等
- 本人用该软件导出Markdown为了方便快速迁移至Obsidian
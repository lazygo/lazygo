package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed examples
var project embed.FS

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("lazygo [create|init] PackageName [ProjectName]\n")
		return
	}

	var packageName string
	var projectName string

	op := strings.ToLower(os.Args[1])
	switch op {
	case "create":
		if len(os.Args) < 4 {
			fmt.Printf("lazygo create PackageName ProjectName\n")
			return
		}
		packageName = os.Args[2]
		projectName = os.Args[3]

		_, err := os.Stat(projectName)
		if !os.IsNotExist(err) {
			fmt.Printf("%s 已存在\n", projectName)
			return
		}
		err = os.Mkdir(projectName, os.ModePerm)
		if err != nil {
			fmt.Printf("%s 创建失败\n", projectName)
			return
		}

	case "init":
		if len(os.Args) < 3 {
			fmt.Printf("lazygo create PackageName ProjectName\n")
			return
		}
		packageName = os.Args[2]
		empty, err := DirIsEmpty(".")
		if err != nil {

		}
		if !empty {
			fmt.Printf("current dir not empty\n")
			return
		}
	default:
		fmt.Printf("lazygo create|init PackageName [ProjectName]\n")
		return
	}

	if projectName == "" {
		fmt.Printf("lazygo create|init PackageName [ProjectName]\n")
		return
	}

	project, err := fs.Sub(project, "examples")
	if err != nil {
		panic(err) // unexpected or a typo
	}
	fs.WalkDir(project, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." || path == ".." {
			return nil
		}
		fmt.Println(path)
		projectPath := strings.TrimRight(filepath.Join(projectName, path), "_")
		if d.IsDir() {
			err = os.Mkdir(projectPath, os.ModePerm)
			if err != nil {
				fmt.Printf("%s 创建失败\n", projectPath)
				return err
			}
			return nil
		}

		data, err := fs.ReadFile(project, path)
		if err != nil {
			fmt.Printf("%s 读取失败\n", projectPath)
			return err
		}

		content := strings.Replace(string(data), "github.com/lazygo/lazygo/examples", packageName, -1)

		err = os.WriteFile(projectPath, []byte(content), os.ModePerm)
		if err != nil {
			fmt.Printf("%s 写入失败\n", projectPath)
			return err
		}
		return nil
	})
}

func DirIsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

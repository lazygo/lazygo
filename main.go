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

// 项目模板模式
const (
	modeExecute = "execute"
	modeDylib   = "dylib"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("lazygo [create|init] PackageName [ProjectName] [execute|dylib]\n")
		return
	}

	var packageName string
	var projectName string
	var mode string

	op := strings.ToLower(os.Args[1])
	args := os.Args[2:]
	switch op {
	case "create":
		if len(args) < 2 {
			fmt.Printf("lazygo create PackageName ProjectName [execute|dylib]\n")
			return
		}
		packageName = args[0]
		projectName = args[1]
		args = args[2:]

		_, err := os.Stat(projectName)
		if !os.IsNotExist(err) {
			fmt.Printf("目录 %s 已存在\n", projectName)
			return
		}
		err = os.Mkdir(projectName, os.ModePerm)
		if err != nil {
			fmt.Printf("目录 %s 创建失败\n", projectName)
			return
		}

	case "init":
		if len(args) < 1 {
			fmt.Printf("lazygo init PackageName [execute|dylib]\n")
			return
		}
		packageName = args[0]
		projectName = "."
		args = args[1:]
		empty, err := DirIsEmpty(".")
		if err != nil {
			fmt.Printf("检查目录失败: %v\n", err)
			return
		}
		if !empty {
			fmt.Printf("current dir not empty\n")
			return
		}
	default:
		fmt.Printf("lazygo create|init PackageName [ProjectName] [execute|dylib]\n")
		return
	}

	mode = modeExecute
	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case modeExecute, "exe":
			mode = modeExecute
		case modeDylib:
			mode = modeDylib
		default:
			fmt.Printf("未知模式 %s，可选 execute（默认）/ dylib\n", args[0])
			return
		}
	}

	project, err := fs.Sub(project, "examples")
	if err != nil {
		panic(err) // unexpected or a typo
	}
	fs.WalkDir(project, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		// 按模式过滤模板子树：execute 释放完整应用（跳过dylib），dylib 只释放dylib目录且铺平到项目根
		rel := path
		switch mode {
		case modeDylib:
			if path == "dylib" {
				return nil
			}
			if !strings.HasPrefix(path, "dylib/") {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			rel = strings.TrimPrefix(path, "dylib/")
		case modeExecute:
			if path == "dylib" || strings.HasPrefix(path, "dylib/") {
				if d.IsDir() && path == "dylib" {
					return fs.SkipDir
				}
				return nil
			}
			// 可执行入口位于exe目录，释放到项目根
			if path == "exe" {
				return nil
			}
			if strings.HasPrefix(path, "exe/") {
				rel = strings.TrimPrefix(path, "exe/")
			}
		}

		projectPath := filepath.Join(projectName, rel)
		if strings.HasSuffix(projectPath, "go.mod_") || strings.HasSuffix(projectPath, "go.sum_") {
			projectPath = strings.TrimRight(projectPath, "_")
		}
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

	fmt.Printf("项目创建完成（模式: %s）\n", mode)
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

package main

import (
	//"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/koki-develop/go-fzf"
)

/*
func readPathList(fileName string) []string {
	fp, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	var pathList []string
	for scanner.Scan() {
		pathList = append(pathList, scanner.Text())
	}
	return pathList
}
*/

func cmdSel(ext string) (string, error) {
	var prog string
	switch ext {
	case ".txt":
		prog = "notepad"
	case ".md", ".markdown":
		prog = "start"
	case ".ppt", ".pptx":
		prog = "powerpnt"
	case ".xls", "xlsx", "xlsm":
		prog = "xls"
	default:
		return "", errors.New("this file extension is undefined")
	}
	return prog, nil
}

func main() {

	//items := readPathList("test.txt")
	var items []string

	//TODO: support specific path
	dirPath := "./"

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			items = append(items, path)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, p := range items {
		fmt.Println(p)
	}

	f, err := fzf.New()
	if err != nil {
		log.Fatal(err)
	}

	idxs, err := f.Find(items, func(i int) string { return items[i] })
	if err != nil {
		log.Fatal(err)
	}
	
	//TODO: support multi file select
	var ext string
	var target string
	for _, i := range idxs {
		ext = filepath.Ext(items[i])
		target = items[i]
	}

	prog, err := cmdSel(ext)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(prog, target)

	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
}

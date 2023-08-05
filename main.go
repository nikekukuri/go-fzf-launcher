package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/koki-develop/go-fzf"
)

const (
	FILE = iota
	DIR
)

type Config struct {
	Command Command `json:"command"`
}

type Command struct {
	PowerPoint string `json:"powerpoint"`
	Excel      string `json:"excel"`
	Text       string `json:"text"`
	Markdown   string `json:"markdown"`
	Dir        string `json:"dir"`
}

type Target struct {
	Path string
	Kind int
}

func cmdSel(ext string, cmd Command) (string, error) {
	var prog string
	switch ext {
	case ".txt":
		prog = cmd.Text
	case ".md", ".markdown":
		prog = cmd.Markdown
	case ".ppt", ".pptx":
		prog = cmd.PowerPoint
	case ".xls", "xlsx", "xlsm":
		prog = cmd.Excel
	default:
		return "", errors.New("this file extension is undefined")
	}
	return prog, nil
}

func getOldItems() ([]Target, error) {
	cmd := exec.Command("powershell", "Get-ChildItem", "$env:APPDATA\\Microsoft\\Windows\\Recent\\*", "-Name")
	
	var items []Target
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("ERROR: get oldfile list command execution: ", err)
		return items, err
	}

	output := stdout.String()

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		fileInfo, err := os.Stat(line)
		if err != nil {
			continue
		}

		if fileInfo.IsDir() {
			items = append(items, Target{
				Path: line,
				Kind: DIR,
			})
		} else {
			items = append(items, Target{
				Path: line,
				Kind: FILE,
			})
		}
	}

	return items, nil
}

func getItems(isOldFile bool) ([]Target, error) {
	var dirPath string
	if isOldFile {
		// search oldfile
		items, err := getOldItems()
		if err != nil {
			log.Fatal(err)
		}
		return items, nil
	} else {
		// seach from path
		if len(os.Args) == 1 {
			// if not specify path, current directory.
			dirPath = "./"
		} else {
			// if specified path, the path store in `dirPath` variable
			dirPath = os.Args[1]
		}
	}

	var items []Target
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil { 
			return err
		}

		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			items = append(items, Target{
				Path: path,
				Kind: FILE,
			})
		} else {
			items = append(items, Target{
				Path: path,
				Kind: DIR,
			})
		}

		return nil
	})

	if err != nil {
		return items, err
	}
	
	return items, nil
}

func main() {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("ERROR: cannot get Home Directory path.", err)
	}

	configFilePath := homeDir + "/.fzl_config.json"

	jsonData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		fmt.Println("Not found `~/.fzl_config.json` config file.")
	}

	var cfg Config
	err = json.Unmarshal(jsonData, &cfg)
	if err != nil {
		fmt.Println("ERROR: cannot json parse")
		log.Fatal(err)
	}


	// parse command line argument
	oldFileFlag := flag.Bool("oldfile", false, "oldfiles flag")
	shortOldFileFlag := flag.Bool("o", false, "oldfiles(short) flag")
	flag.Parse()
	isOldFile := *oldFileFlag || *shortOldFileFlag

	items, err := getItems(isOldFile)
	
	if err != nil {
		log.Fatal(err)
	}

	f, err := fzf.New()
	if err != nil {
		log.Fatal(err)
	}

	idxs, err := f.Find(items, func(i int) string { return items[i].Path })
	if err != nil {
		log.Fatal(err)
	}
	
	var ext, target, prog string
	//TODO: support multi file select
	for _, i := range idxs {
		if items[i].Kind == FILE {
			ext = filepath.Ext(items[i].Path)
			target = items[i].Path
			prog, err = cmdSel(ext, cfg.Command)
		} else {
			ext = ""
			target = items[i].Path
			prog = cfg.Command.Dir
		}
	}

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

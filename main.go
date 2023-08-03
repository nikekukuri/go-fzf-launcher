package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("ERROR: cannot get Home Directory path.", err)
	}

	configFilePath := homeDir + "/.fla_config.json"

	jsonData, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		fmt.Println("not existence `config.json` file, default values are applied.")
	}

	var cfg Config
	err = json.Unmarshal(jsonData, &cfg)
	if err != nil {
		fmt.Println("ERROR: cannot json parse")
		log.Fatal(err)
	}

	fmt.Println("powerpoint: ", cfg.Command.PowerPoint)
	fmt.Println("excel: ", cfg.Command.Excel)

	var items []Target

	//TODO: support specific path
	dirPath := "./"

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err
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
			prog, err = cmdSel(ext)
		} else {
			ext = ""
			target = items[i].Path
			prog = "cd"
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	cmdExec := exec.Command(prog, target)

	output, err := cmdExec.Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(output))
}

package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var rootDir string

func main() {
	var skipPreWatch bool

	flag.StringVar(&rootDir, "dir", "nas-data/data/downloads", "Directory to start watching")
	flag.BoolVar(&skipPreWatch, "spw", false, "Skip the pre watch")

	if !skipPreWatch {
		handlePreWatch()
	}
	
	handleWatcher()

	<-make(chan struct{})
	// Block main goroutine forever.
}

func handleWatcher() {
	log.Println("Starting Watch instance for " + rootDir)
	watcher, err := fsnotify.NewWatcher()
	if hasError(err) {
		panic("watcher setup failed")
	}

	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if hasError(err) {
			panic("watcher setup failed")
		}
	}(watcher)

	go listenForEvents(watcher)

	err = watcher.Add(rootDir)
	if hasError(err) {
		panic("watcher setup failed")
	}
}

func listenForEvents(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			handleEvent(event)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			hasError(err)
		}
	}
}

func handleEvent(event fsnotify.Event) {
	log.Println("Event captured: " + event.Name)

	if !event.Has(fsnotify.Create) {
		return
	}

	log.Println("Created event captured: " + event.Name)

	file, err := os.Open(event.Name)
	if hasError(err) {
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if hasError(err) {
			panic("Cannot close file handle")
		}
	}(file)

	stat, err := file.Stat()

	if hasError(err) {
		return
	}

	if stat.IsDir() {
		handleDir(filepath.Join(rootDir, stat.Name()))
		return
	}

	if !isArchive(file.Name()) {
		return
	}

	handleCommandExecution(event.Name, rootDir)
}

func handlePreWatch() {
	dirs, err := os.ReadDir(rootDir)
	if hasError(err) {
		return
	}

	for _, dir := range dirs {
		if dir.IsDir() {

			if hasError(err) {
				continue
			}

			handleDir(filepath.Join(rootDir, dir.Name()))
		}
	}
}

func handleDir(dir string) {
	files, err := os.ReadDir(dir)

	if hasError(err) {
		return
	}

	for _, file := range files {
		if !isArchive(file.Name()) {
			continue
		}

		handleCommandExecution(filepath.Join(dir, file.Name()), dir)
	}
}

func handleCommandExecution(fileName string, outDir string) {
	cmd := exec.Command("/usr/bin/7z", "x", fileName, "-o"+outDir, "-aos")
	result, err := cmd.CombinedOutput()

	log.Println(string(result))

	if hasError(err) {
		return
	}
}

func isArchive(fileName string) bool {
	return strings.Contains(fileName, ".rar")
}

func hasError(err error) bool {
	if err != nil {
		log.Println("error:", err)
		return true
	}
	return false
}

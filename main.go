package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

const rootDir = "/nas-data/data/downloads"

func main() {
	handlePreWatch()
	handleWatcher()

	<-make(chan struct{})
	// Block main goroutine forever.
}

func handleWatcher() {
	log.Println("Starting Watch instance for " + rootDir)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(watcher)

	go listenForEvents(watcher)

	err = watcher.Add(rootDir)
	if err != nil {
		log.Fatal(err)
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
			log.Println("error:", err)
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
	if isError(err) {
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if isError(err) {
			panic("Cannot close file handle")
		}
	}(file)

	stat, err := file.Stat()

	if isError(err) {
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
	if isError(err) {
		return
	}

	for _, dir := range dirs {
		if dir.IsDir() {

			if isError(err) {
				continue
			}

			handleDir(filepath.Join(rootDir, dir.Name()))
		}
	}
}

func handleDir(dir string) {
	files, err := os.ReadDir(dir)

	if isError(err) {
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

	if isError(err) {
		return
	}
}

func isArchive(fileName string) bool {
	if strings.Contains(fileName, ".rar") {
		return true
	}

	return false
}

func isError(err error) bool {
	if err != nil {
		log.Println("error:", err)
		return true
	}
	return false
}

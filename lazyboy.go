package main

import (
	"lazyboy/queue"
	"log"
	"os"
	"path"
	"time"
)

func tick(pipeBasePath string) {
	log.Println("PIPE", pipeBasePath)
	dirs, err := os.ReadDir(pipeBasePath)
	if err != nil {
		log.Println(err)
		return
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			go pipe(path.Join(pipeBasePath, dir.Name()))
		}
	}
}
func pipe(pipePath string) {
	log.Println("PIPE", pipePath)
	// 1. SETUP
	// load config from path
	pipeline, err := queue.NewPipelineFromConfigPath(path.Join(pipePath, "config.json"))
	if err != nil {
		log.Println(err)
		return
	}

	// 2. TAKE
	taken := pipeline.Take()

	// 3. MERGE DATA
	for _, t := range taken {
		request, err := pipeline.BuildRequest(t)
		if err != nil {
			return
		}
		log.Println(request)
	}

}
func run(pipeBasePath string) {
	ticker := time.NewTicker(time.Second)
	//go func() {
	for t := range ticker.C {
		log.Println("Tick at", t)
		tick(pipeBasePath)
	}
	//}()
}
func main() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	run(path.Join(wd, "testcases", "pipelines"))
}

package main

import (
	"context"
	"encoding/json"
	"lazyboy/queue"
	"log"
	"os"
	"path"
	"time"
)

func tick(pipeBasePath string) {
	ctx := context.Background()
	log.Println("PIPE", pipeBasePath)
	dirs, err := os.ReadDir(pipeBasePath)
	if err != nil {
		log.Println(err)
		return
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			go pipe(ctx, path.Join(pipeBasePath, dir.Name()))
		}
	}
}
func pipe(ctx context.Context, pipePath string) {
	log.Println("PIPE", pipePath)
	// 1. SETUP
	// load config from path
	pipeline, err := queue.NewPipelineFromConfigPath(path.Join(pipePath, "config.json"))
	if err != nil {
		log.Println(err)
		return
	}

	if !pipeline.IsActive(time.Now()) {
		log.Println("Not active")
		return
	}

	// 2. TAKE
	taken := pipeline.Take()

	// 3. MERGE DATA
	for _, t := range taken {
		var tobj interface{}
		err := json.Unmarshal(t, &tobj)
		if err != nil {
			log.Println("Invalid line", err)
			continue
		}
		req, err := queue.NewReqFromPipeline(pipeline, tobj)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(req)

		res := req.Run(ctx, pipeline)
		if res.Err != "" {
			log.Println("RUN ERR:", res.Err)
		}

		log.Printf("Res : %#v", res)

		resobj, err := res.ParseResponse(pipeline)
		if err != nil {
			return
		}

		resout, err := json.Marshal(resobj)
		if err != nil {
			return
		}

		log.Printf("Log : %v", string(resout))
	}

}
func run(pipeBasePath string) {
	debug := os.Getenv("LAZYBOY_DEBUG")
	var ticker *time.Ticker
	if debug != "" {
		ticker = time.NewTicker(time.Second)
	} else {
		ticker = time.NewTicker(time.Minute)
	}
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
	run(path.Join(wd, "pipelines"))
}

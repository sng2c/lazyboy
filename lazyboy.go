package main

import (
	"context"
	"encoding/json"
	"lazyboy/queue"

	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

func tick(pipeBasePath string) {
	ctx := context.Background()
	log.Infof("BasePath : %v", pipeBasePath)
	dirs, err := os.ReadDir(pipeBasePath)
	if err != nil {
		log.Warnf("Reading Dir failed - %v", err)
		return
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			go pipe(ctx, path.Join(pipeBasePath, dir.Name()))
		}
	}
}

func pipe(ctx context.Context, pipePath string) {
	logger := log.WithField("PipePath", pipePath)
	logger.Info("Begin Pipe")
	// 1. SETUP
	// load config from path
	pipeline, err := queue.NewPipelineFromConfigPath(path.Join(pipePath, "config.json"))
	if err != nil {
		logger.Debugf("Can not load config.json in PipePath - %v", err)
		return
	}

	if !pipeline.IsActive(time.Now()) {
		logger.Warnf("Not active. ActiveTime : %v", pipeline.ActiveTime)
		return
	}

	// 2. TAKE
	taken := pipeline.Take()

	if len(taken) == 0 {
		logger.Warnf("Empty")
	} else {
		// 3. MERGE DATA
		for i, t := range taken {
			var tobj interface{}
			err := json.Unmarshal(t, &tobj)
			if err != nil {
				logger.Warnf("Invalid line #%v - %v", i, err)
				continue
			}
			req, err := queue.NewReqFromPipeline(pipeline, tobj)
			if err != nil {
				logger.Warnf("Building Req failed %v", err)
				continue
			}
			logger.Debugf("Req : %#v", req)

			res := req.Run(ctx, pipeline)
			if res.Err != "" {
				logger.Println("RUN ERR:", res.Err)
			}

			logger.Debugf("Res : %#v", res)

			resout, err := res.ParseResponse(pipeline)
			if err != nil {
				logger.Warnf("Rendering output failed - %v", err)
				return
			}

			logger.Debugf("Output : %v", string(resout))
		}
	}

}
func run(pipeBasePath string) {
	debug := os.Getenv("LAZYBOY_DEBUG")
	var ticker *time.Ticker
	if debug != "" {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
		ticker = time.NewTicker(time.Second)
	} else {
		ticker = time.NewTicker(time.Minute)
	}
	//go func() {
	for t := range ticker.C {
		log.Infof("Tick at %v", t)
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

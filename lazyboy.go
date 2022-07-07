package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/pool.v3"
	"lazyboy/queue"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"
)

type Work struct {
	Ctx       context.Context
	Pipe      *queue.Pipeline
	Req       *queue.Req
	Res       *queue.Res
	Index     int
	UniqueKey interface{}
}

func runHttpWorkFunc(work *Work) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		work.Ctx, _ = context.WithTimeout(context.Background(), time.Second*30)
		res := work.Req.Run(work.Ctx, work.Pipe)
		work.Res = res
		return work, nil // everything ok, send nil, error if not
	}
}
func proc(ctx context.Context, wg *sync.WaitGroup, pipePath string) {

	logger := logrus.WithContext(ctx)

	wg.Add(1)
	defer wg.Done()

	logger.Debugf("Begin Proc with %v", pipePath)
	// 1. SETUP
	// load config from path
	pipe, err := queue.NewPipelineFromConfigPath(path.Join(pipePath, "config.json"))
	if err != nil {
		logger.Debugf("Can not load config.json in PipePath - %v", err)
		return
	}

	outlogger := logrus.New()
	file, err := os.OpenFile(pipe.OutputAbsPath(), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		logger.Warnf("Cannot Generate output file - %v", err)
		return
	}
	outlogger.SetFormatter(&logrus.JSONFormatter{})
	outlogger.SetOutput(file)

	logger = logrus.WithFields(logrus.Fields{"Pipename": pipe.GetName()})

	if !pipe.IsActive(time.Now()) {
		logger.Warnf("Not active. ActiveTime : %v", pipe.ActiveTime)
		return
	}

	workers := pipe.Workers
	if workers < 1 {
		logger.Warn("Workers need to be greater than 0")
		workers = 1
	}

	// 2. TAKE
	taken := pipe.Take()

	if len(taken) == 0 {
		logger.Warnf("Empty")
	} else {

		p := pool.NewLimited(uint(workers))
		defer p.Close()

		batch := p.Batch()

		// 3. MERGE DATA
		for i, t := range taken {
			var takenObj interface{}
			err := json.Unmarshal(t, &takenObj)
			if err != nil {
				logger.Warnf("Invalid line #%v - %v", i+1, err)
				outlogger.WithError(err).WithField("UniqueKey", nil).Errorln("error")
				continue
			}

			uniqueKey, err := pipe.GetUniqueKey(takenObj)
			if err != nil {
				logger.Warnf("No uniqueKey '%v' in data - %v", pipe.UniqueKey, err)
				outlogger.WithError(err).WithField("UniqueKey", nil).WithField("data", takenObj).Errorln("error")
				continue
			}

			logger.Infof("proc %v (%v/%v)", uniqueKey, i+1, len(taken))

			req, err := queue.NewReqFromPipeline(pipe, takenObj)
			if err != nil {
				logger.Warnf("Building Req failed %v", err)
				outlogger.WithError(err).WithField("UniqueKey", uniqueKey).Errorln("error")
				continue
			}
			logger.Debugf("Req : %#v", req)

			batch.Queue(runHttpWorkFunc(&Work{
				Ctx:       ctx,
				Pipe:      pipe,
				Req:       req,
				Index:     i,
				UniqueKey: uniqueKey,
			}))

		}

		batch.QueueComplete()

		for workRes := range batch.Results() {
			work := workRes.Value().(*Work)
			ctx2 := work.Ctx
			pipe2 := work.Pipe
			res := work.Res
			i := work.Index
			uniqueKey := work.UniqueKey

			if res.Err != "" {
				logger.Warnf("Http Error: %v", res.Err)
				outlogger.WithError(errors.New(res.Err)).WithField("UniqueKey", uniqueKey).Errorln("error")
				continue
			}

			resout, err := res.BuildOutput(pipe2)
			if err != nil {
				logger.Warnf("Building output failed - %v", err)
				outlogger.WithError(err).WithField("UniqueKey", uniqueKey).Errorln("error")
				continue
			}
			outlogger.WithField("UniqueKey", uniqueKey).WithField("result", resout).Println("ok")

			logger.Infof("done %v (%v/%v)", uniqueKey, i+1, len(taken))

			if ctx2.Value("debug") != nil {
				time.Sleep(time.Second)
			}

		}
	}

	logger.Debug("End Proc")

}

func tick(ctx context.Context, wg *sync.WaitGroup, pipeBasePath string) {
	logger := logrus.WithContext(ctx)
	logger.Debug("Tick Begin")
	dirs, err := os.ReadDir(pipeBasePath)
	if err != nil {
		logger.Warnf("Reading Dir failed - %v", err)
		return
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			ctxQueue := context.WithValue(ctx, "queuePath", dir.Name())
			go proc(ctxQueue, wg, path.Join(pipeBasePath, dir.Name()))
		}
	}
	logger.Debug("Tick End")
}

func run(ctx context.Context, pipeBasePath string) {
	ctxRun := context.WithValue(ctx, "basePath", pipeBasePath)
	logger := logrus.WithContext(ctxRun)
	var ticker *time.Ticker
	var tickerInterval time.Duration = time.Minute
	if ctx.Value("debug") != nil {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetReportCaller(true)
		tickerInterval = time.Second
	} else {
		logrus.SetLevel(logrus.InfoLevel)
		tickerInterval = time.Minute
	}
	ticker = time.NewTicker(tickerInterval)
	logger.Infof("Ticker Begins interval %v", tickerInterval)
	//go func() {
	wg := &sync.WaitGroup{}

SELECT:
	for {
		select {
		case t := <-ticker.C:
			ctxTick := context.WithValue(ctxRun, "tickAt", t)
			logger.Infof("Tick at %v", t)
			tick(ctxTick, wg, pipeBasePath)
		case <-ctx.Done():
			ticker.Stop()
			break SELECT
		}
	}

	logger.Info("Waiting processing...")
	wg.Wait()
	logger.Info("All Done")
}

func main() {

	// TODO line truncation 할것
	// TODO Logrotate

	var queueBaseDir string
	flag.StringVar(&queueBaseDir, "d", "queuebase", "Queue base directory")
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		return
	}

	ctx := context.Background()

	debug := os.Getenv("LAZYBOY_DEBUG")
	if debug != "" {
		ctx = context.WithValue(ctx, "debug", true)
	}

	ctx, cancelFunc := context.WithCancel(ctx)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancelFunc()
	}()
	run(ctx, path.Join(wd, queueBaseDir))
}

package queue

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"strings"
)

type FileQueue struct {
	QueuePath     string
	FileQueueName string
	Pos           FileQueuePos
}
type FileQueuePos struct {
	Offset    int64
	LastError string `json:",omitempty"`
}

var ErrNoData = errors.New("no more data")

func NewFileQueue(queuePath, queueName string) (*FileQueue, error) {

	fqPath := path.Join(queuePath, queueName)
	logger := logrus.WithFields(logrus.Fields{"ctx": "NewFileQueue", "path": fqPath})
	fq := FileQueue{
		QueuePath:     queuePath,
		FileQueueName: queueName,
		Pos: FileQueuePos{
			Offset:    0,
			LastError: "",
		},
	}

	_, err := os.Stat(fqPath)
	if err != nil {
		return nil, err
	}

	// Init pos
	posPath := path.Join(fq.QueuePath, fq.FileQueueName+".pos")
	posData, err := os.ReadFile(posPath)
	if os.IsNotExist(err) {
		logger.Debugf("Use default Pos. %v", err)
		err := fq.SyncPos()
		if err != nil {
			return nil, err
		}
	} else {
		err = json.Unmarshal(posData, &fq.Pos)
		if err != nil {
			logger.Debugf("Use default Pos. %v", err)
			err := fq.SyncPos()
			if err != nil {
				return nil, err
			}
		}
	}

	return &fq, nil
}

func (fq *FileQueue) IsEOF() bool {
	subqPath := path.Join(fq.QueuePath, fq.FileQueueName)
	stat, err := os.Stat(subqPath)
	if os.IsNotExist(err) {
		return true
	}
	if stat.Size() <= fq.Pos.Offset {
		return true
	}
	return false
}

func (fq *FileQueue) SyncPos() error {
	marshaled, _ := json.Marshal(fq.Pos)
	posPath := path.Join(fq.QueuePath, fq.FileQueueName+".pos")
	err := os.WriteFile(posPath, marshaled, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (fq *FileQueue) Take(n int) [][]byte {
	logger := logrus.WithFields(logrus.Fields{"ctx": "queue/FileQueue.Take", "path": path.Join(fq.QueuePath, fq.FileQueueName)})
	file, err := os.OpenFile(path.Join(fq.QueuePath, fq.FileQueueName), os.O_RDONLY, 0)
	if err != nil {
		fq.Pos.LastError = err.Error()
		return nil
	}
	defer file.Close()
	_, err = file.Seek(fq.Pos.Offset, io.SeekStart)
	if err != nil {
		fq.Pos.LastError = err.Error()
		return nil
	}
	var taken = make([][]byte, 0)

	rd := bufio.NewReader(file)

	// Take후에는 반드시 싱크
	defer func(q *FileQueue) {
		logger.Debug("Sync", fq)
		err := q.SyncPos()
		if err != nil {
			logger.Debug(err)
		}
	}(fq)

	var hasRead int64
	for i := 0; i < n; i++ {
		bytes, err := rd.ReadBytes('\n')
		if err != nil {

			if errors.Is(io.EOF, err) { // err이 있더라도 bytes는 채워져있음
				logger.Debug("Take EOF")
			} else {
				fq.Pos.LastError = err.Error()
				logger.Debug(err)
				return nil
			}
		}
		hasRead += int64(len(bytes))
		if len(bytes) > 0 && bytes[len(bytes)-1] == '\n' {
			bytes = bytes[:len(bytes)-1]
		}
		if len(bytes) > 0 && bytes[len(bytes)-1] == '\r' {
			bytes = bytes[:len(bytes)-1]
		}
		if len(bytes) > 0 {
			taken = append(taken, bytes)
		}
		if fq.Pos.LastError != "" {
			break
		}
	}
	fq.Pos.Offset += hasRead

	return taken
}

func OfferFileQueue(queuePath string) (*FileQueue, error) {
	logger := logrus.WithField("path", queuePath)
	dirs, err := os.ReadDir(queuePath)
	if err != nil {
		return nil, err
	}
	// 디렉토리와 "_" 로 시작하는 파일 제거
	var targets []*FileQueue
	for _, d := range dirs {
		if d.IsDir() {
			continue
		}
		if !strings.HasSuffix(d.Name(), ".jsonl") {
			continue
		}
		// 정합성 체크를 여기서 끝낸다.
		fq, err := NewFileQueue(queuePath, d.Name())
		if err != nil {
			logger.Debug("Skip by Error", err)
			continue
		}
		if fq.Pos.LastError != "" {
			logger.Debug("Skip by LastError", fq.Pos.LastError)
			continue
		}
		if fq.IsEOF() {
			logger.WithField("dir", d.Name()).Debug("Skip by EOF")
			continue
		}
		targets = append(targets, fq)
	}
	// os.ReadDir 에서 이미 정렬이 되어있어서 정렬코드 생략.
	//sort.SliceStable(targets, func(i, j int) bool {
	//	compare := strings.Compare(targets[i].Name(), targets[j].Name())
	//	return compare == -1
	//})
	if len(targets) > 0 {
		return targets[0], nil
	} else {
		return nil, ErrNoData
	}
}

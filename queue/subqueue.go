package queue

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type SubQueue struct {
	QueuePath    string
	SubQueueName string
	Pos          SubQueuePos
}
type SubQueuePos struct {
	Offset    int64
	HasError  bool
	LastError string
}

func NewSubQueue(queuePath, queueName string) (*SubQueue, error) {
	subq := SubQueue{
		QueuePath:    queuePath,
		SubQueueName: queueName,
		Pos: SubQueuePos{
			Offset:    0,
			HasError:  false,
			LastError: "",
		},
	}

	subqPath := path.Join(subq.QueuePath, subq.SubQueueName)
	_, err := os.Stat(subqPath)
	if os.IsNotExist(err) {
		log.Println("No subq", err)
	}

	// Init pos
	posPath := path.Join(subq.QueuePath, subq.SubQueueName+".pos")
	posData, err := os.ReadFile(posPath)
	if os.IsNotExist(err) {
		log.Println("Use default Pos. " + err.Error())
		subq.SyncPos()
	} else {
		err = json.Unmarshal(posData, &subq.Pos)
		if err != nil {
			log.Println("Use default Pos. " + err.Error())
			subq.SyncPos()
		}
	}

	return &subq, nil
}

func (subq *SubQueue) IsEOF() bool {
	subqPath := path.Join(subq.QueuePath, subq.SubQueueName)
	stat, err := os.Stat(subqPath)
	if os.IsNotExist(err) {
		return true
	}
	if stat.Size() <= subq.Pos.Offset {
		return true
	}
	return false
}

func (subq *SubQueue) SyncPos() error {
	marshaled, _ := json.Marshal(subq.Pos)
	posPath := path.Join(subq.QueuePath, subq.SubQueueName+".pos")
	err := os.WriteFile(posPath, marshaled, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (subq *SubQueue) Take(n int) [][]byte {
	file, err := os.OpenFile(path.Join(subq.QueuePath, subq.SubQueueName), os.O_RDONLY, 0)
	if err != nil {
		subq.Pos.HasError = true
		subq.Pos.LastError = err.Error()
		return nil
	}
	defer file.Close()
	_, err = file.Seek(subq.Pos.Offset, io.SeekStart)
	if err != nil {
		subq.Pos.HasError = true
		subq.Pos.LastError = err.Error()
		return nil
	}
	var taken = make([][]byte, 0)

	rd := bufio.NewReader(file)

	// Take후에는 반드시 싱크
	defer func(q *SubQueue) {
		log.Println("Sync", subq)
		err := q.SyncPos()
		if err != nil {
			log.Println(err)
		}
	}(subq)

	var hasRead int64
	for i := 0; i < n; i++ {
		bytes, err := rd.ReadBytes('\n')
		if err != nil {

			if errors.Is(io.EOF, err) { // err이 있더라도 bytes는 채워져있음
				log.Println("Take EOF")
			} else {
				subq.Pos.HasError = true
				subq.Pos.LastError = err.Error()
				log.Println(err)
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
		if subq.Pos.HasError {
			break
		}
	}
	subq.Pos.Offset += hasRead

	return taken
}

func OfferSubQueue(queuePath string) (*SubQueue, error) {
	dirs, err := os.ReadDir(queuePath)
	if err != nil {
		return nil, err
	}
	// 디렉토리와 "_" 로 시작하는 파일 제거
	var targets []*SubQueue
	for _, d := range dirs {
		if d.IsDir() {
			continue
		}
		if !strings.HasSuffix(d.Name(), ".jsonl") {
			continue
		}
		// 정합성 체크를 여기서 끝낸다.
		subq, err := NewSubQueue(queuePath, d.Name())
		if err != nil {
			continue
		}
		if subq.Pos.HasError {
			log.Println("Skip by LastError", subq.Pos.LastError)
			continue
		}
		if subq.IsEOF() {
			log.Println("Skip by EOF")
			continue
		}
		targets = append(targets, subq)
	}
	// os.ReadDir 에서 이미 정렬이 되어있어서 정렬코드 생략.
	//sort.SliceStable(targets, func(i, j int) bool {
	//	compare := strings.Compare(targets[i].Name(), targets[j].Name())
	//	return compare == -1
	//})
	if len(targets) > 0 {
		return targets[0], nil
	} else {
		return nil, os.ErrNotExist
	}
}

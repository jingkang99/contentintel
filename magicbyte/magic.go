package magicbyte

import (
	"context"
	"fmt"
	"runtime"
)

type job struct {
	input      []byte
	reference  FileType
	resultChan chan *FileType
}

// ErrUnknown infers the file type cannot be determined by the provided magic bytes
var ErrUnknown = fmt.Errorf("unknown file type")

// Provide at least the first 1024 bytes. A magic.ErrUnknown will be returned if unknown
func Lookup(bytes []byte) (*FileType, error) {

	// use all available cores
	workerCount := runtime.GOMAXPROCS(0)
	workChan := make(chan job)
	resultChan := make(chan *FileType)
	
	//fmt.Printf("  working cores: %X\n", workerCount)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// spawn workers
	for i := 0; i < workerCount; i++ {
		go worker(ctx, workChan)
	}

	awaiting := len(types)

	// queue work
	go func() {
		for _, t := range types {
			select {
			case <-ctx.Done():
				return
			case workChan <- job{
				input:      bytes,
				reference:  t,
				resultChan: resultChan,
			}:
			}
		}

	}()

	for {
		result := <-resultChan
		if result != nil {
			return result, nil
		}
		awaiting--
		if awaiting <= 0 {
			break
		}
	}

	return nil, ErrUnknown
}

func worker(ctx context.Context, work chan job) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-work:
			select {
			case <-ctx.Done():
				return
			case job.resultChan <- job.reference.check(job.input, 0):
			}
		}
	}
}

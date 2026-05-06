package pool

import (
	"errors"
	"math/rand/v2"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type TaskHandlerFunc[T any] func(task T)

type WorkerPool[T any] struct {
	handlerFunc        TaskHandlerFunc[T]
	idleWorkerLifetime time.Duration
	numShards          int
	shards             []*poolShard[T]
	mutex              spinLocker
	started            bool
	stopped            bool
	_                  [56]byte
	spawnedWorkers     atomic.Uint64
}

type workerInstance[T any] struct {
	taskChan  chan T
	shard     *poolShard[T]
	lastUsed  time.Time
	isDeleted bool
	_         [16]byte
}

type poolShard[T any] struct {
	wp             *WorkerPool[T]
	workerCache    sync.Pool
	idleWorkerList []*workerInstance[T]
	idleWorker1    *workerInstance[T]
	idleWorker2    *workerInstance[T]
	mutex          spinLocker
	stopped        bool
}

const (
	defaultIdleWorkerLifetime = time.Second
	maxShards                 = 128
)

func NewWorkerPool[T any](handlerFunc TaskHandlerFunc[T]) *WorkerPool[T] {
	wp := &WorkerPool[T]{
		handlerFunc:        handlerFunc,
		idleWorkerLifetime: defaultIdleWorkerLifetime,
		numShards:          1,
	}

	wp.SetNumShards(runtime.GOMAXPROCS(0))
	return wp
}

func (wp *WorkerPool[T]) SetNumShards(numShards int) {
	if numShards <= 1 {
		numShards = 1
	}

	if numShards > maxShards {
		numShards = maxShards
	}

	wp.numShards = numShards
}

func (wp *WorkerPool[T]) SetIdleWorkerLifetime(d time.Duration) {
	wp.idleWorkerLifetime = d
}

func (wp *WorkerPool[T]) GetSpawnedWorkers() int {
	return int(wp.spawnedWorkers.Load())
}

func (wp *WorkerPool[T]) Start() {
	wp.mutex.Lock()
	if !wp.started {
		for i := 0; i < wp.numShards; i++ {
			shard := &poolShard[T]{
				wp: wp,
				workerCache: sync.Pool{
					New: func() any {
						return &workerInstance[T]{
							taskChan: make(chan T),
						}
					},
				},

				idleWorkerList: make([]*workerInstance[T], 0, 2048),
			}
			wp.shards = append(wp.shards, shard)
		}

		wp.started = true
	}
	wp.mutex.Unlock()

	go wp.cleanup()
}

func (wp *WorkerPool[T]) Stop() {
	wp.mutex.Lock()
	if !wp.started {
		wp.mutex.Unlock()
		return
	}

	if !wp.stopped {
		for i := 0; i < wp.numShards; i++ {
			shard := wp.shards[i]
			shard.mutex.Lock()
			shard.stopped = true
			for j := 0; j < len(shard.idleWorkerList); j++ {
				if !shard.idleWorkerList[j].isDeleted {
					shard.idleWorkerList[j].isDeleted = true
					close(shard.idleWorkerList[j].taskChan)
				}
			}
			shard.mutex.Unlock()
		}
	}
	wp.stopped = true
	wp.mutex.Unlock()
}

func (wp *WorkerPool[T]) AddTask(task T) error {
	if !wp.started {
		return errors.New("worker pool must be started first")
	}

	shard := wp.shards[rand.IntN(wp.numShards)] //nolint:gosec
	shard.getWorker(task)

	return nil
}

func (wp *WorkerPool[T]) AddTaskForShard(task T, shardIdx int) error {
	if !wp.started {
		return errors.New("worker pool must be started first")
	}

	shard := wp.shards[shardIdx%wp.numShards]
	shard.getWorker(task)

	return nil
}

func (shard *poolShard[T]) getWorker(task T) (worker *workerInstance[T]) {
	ptr1 := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&shard.idleWorker1)))                                  //nolint:gosec
	if ptr1 != nil && atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&shard.idleWorker1)), ptr1, nil) { //nolint:gosec
		worker = (*workerInstance[T])(ptr1)
		worker.taskChan <- task
		return worker
	}

	ptr2 := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&shard.idleWorker2)))                                  //nolint:gosec
	if ptr2 != nil && atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&shard.idleWorker2)), ptr2, nil) { //nolint:gosec
		worker = (*workerInstance[T])(ptr2)
		worker.taskChan <- task
		return worker
	}

	shard.mutex.Lock()
	iws := len(shard.idleWorkerList)
	if iws > 0 {
		worker = shard.idleWorkerList[iws-1]
		shard.idleWorkerList[iws-1] = nil
		shard.idleWorkerList = shard.idleWorkerList[0 : iws-1]
		shard.mutex.Unlock()
		worker.taskChan <- task
		return worker
	}
	shard.mutex.Unlock()

	worker = shard.workerCache.Get().(*workerInstance[T])
	worker.shard = shard
	go worker.run()

	worker.taskChan <- task
	return worker
}

func (worker *workerInstance[T]) run() {
	shard := worker.shard
	wp := shard.wp
	wp.spawnedWorkers.Add(+1)

	for task := range worker.taskChan {
		wp.handlerFunc(task)
		if !shard.setWorkerIdle(worker) {
			break
		}
	}

	wp.spawnedWorkers.Add(^uint64(0))
	shard.workerCache.Put(worker)
}

func (shard *poolShard[T]) setWorkerIdle(worker *workerInstance[T]) bool {
	worker.lastUsed = time.Now()

	if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&shard.idleWorker2)), nil, unsafe.Pointer(worker)) { //nolint:gosec
		return true
	}
	if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&shard.idleWorker1)), nil, unsafe.Pointer(worker)) { //nolint:gosec
		return true
	}

	worker.shard.mutex.Lock()
	if !worker.shard.stopped {
		worker.shard.idleWorkerList = append(worker.shard.idleWorkerList, worker)
	}
	worker.shard.mutex.Unlock()
	return !worker.shard.stopped
}

func (wp *WorkerPool[T]) cleanup() {
	var toBeCleaned []*workerInstance[T]
	for {
		time.Sleep(wp.idleWorkerLifetime)
		if wp.stopped {
			return
		}

		now := time.Now()
		for i := 0; i < wp.numShards; i++ {
			shard := wp.shards[i]

			shard.mutex.Lock()
			idleWorkerList := shard.idleWorkerList
			iws := len(idleWorkerList)
			s := 0
			j := 0 //nolint

			if iws > 400 {
				s = (iws - 1) / 2
				for s > 0 && now.Sub(idleWorkerList[s].lastUsed) < wp.idleWorkerLifetime {
					s /= 2
				}

				if s == 0 {
					shard.mutex.Unlock()
					continue
				}
			}

			if hasExpiredIdleWorkers(s, iws, now, idleWorkerList, wp) {
				break
			}

			if j == 0 {
				shard.mutex.Unlock()
				continue
			}

			toBeCleaned = append(toBeCleaned[:0], idleWorkerList[0:j]...)

			numMoved := copy(idleWorkerList, idleWorkerList[j:])
			for j = numMoved; j < iws; j++ {
				idleWorkerList[j] = nil
			}
			shard.idleWorkerList = idleWorkerList[:numMoved]
			shard.mutex.Unlock()

			cleanupWorker(toBeCleaned)
		}
	}
}

func hasExpiredIdleWorkers[T any](s, iws int, now time.Time, idleWorkerList []*workerInstance[T], wp *WorkerPool[T]) bool {
	for j := s; j < iws; j++ {
		if now.Sub(idleWorkerList[s].lastUsed) < wp.idleWorkerLifetime {
			return true
		}
	}
	return false
}

func cleanupWorker[T any](toBeCleaned []*workerInstance[T]) {
	for j := range len(toBeCleaned) {
		if !toBeCleaned[j].shard.stopped {
			close(toBeCleaned[j].taskChan)
		}
		toBeCleaned[j] = nil
	}
}

type spinLocker struct {
	lock atomic.Uint64
}

func (s *spinLocker) Lock() {
	schedulerRuns := 1
	for !s.lock.CompareAndSwap(0, 1) {
		for i := 0; i < schedulerRuns; i++ {
			runtime.Gosched()
		}
		if schedulerRuns < 32 {
			schedulerRuns <<= 1
		}
	}
}

func (s *spinLocker) Unlock() {
	s.lock.Store(0)
}

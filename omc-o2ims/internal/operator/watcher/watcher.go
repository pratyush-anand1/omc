package watcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/enrayga/omc-o2ims/internal/operator/resource"
	"github.com/enrayga/omc-o2ims/internal/operator/store"
)

var (
	InfoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarnLogger  = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

type WatcherInterface[T resource.Resource] interface {
	StartWatching() error
	StopWatching() error
	Init(store store.Store[T]) error
}

type WatcherImpl[T resource.Resource] struct {
	storeImpl  store.Store[T]
	isWatching atomic.Bool
	//	isPaused           atomic.Bool
	lastInvocationTime atomic.Int64
	//	heartBeat          atomic.Int64
	stopChan chan struct{}
	wg       sync.WaitGroup
	// mu                 sync.Mutex
}

func (w *WatcherImpl[T]) UpdatelastInvocationTime() {
	t := time.Now()
	w.lastInvocationTime.Store(t.Unix())
}

// GetHeartbeat returns the heartbeat timestamp
func (w *WatcherImpl[T]) GetUpdatelastInvocationTime() time.Time {
	unixTime := w.lastInvocationTime.Load()
	return time.Unix(unixTime, 0)
}

func (w *WatcherImpl[T]) Init(store store.Store[T]) error {
	w.storeImpl = store
	w.stopChan = make(chan struct{})
	return nil
}

func NewWatcherImpl[T resource.Resource](store store.Store[T]) *WatcherImpl[T] {
	return &WatcherImpl[T]{
		storeImpl: store,
		stopChan:  make(chan struct{}),
	}
}

// Start begins the watching process
func (w *WatcherImpl[T]) StartWatching(ctx context.Context) error {
	// Ensure we're not already watching
	if !w.isWatching.CompareAndSwap(false, true) {
		return fmt.Errorf("Watcher is already running!!")
	} else {
		log.Println("Watcher started!!")
	}
	w.wg.Add(1)
	go func() {
		defer func() {
			w.isWatching.Store(false)
			w.wg.Done()
		}()
		for {
			select {
			case <-ctx.Done():
				log.Println("Context Cancelled. Watcher stopped!!")
				return
			case <-w.stopChan:
				return
			default:
				w.UpdatelastInvocationTime()
				if err := w.watch(); err != nil {
					return
				}
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()
	return nil
}

// StopWatching gracefully stops the watcher
func (w *WatcherImpl[T]) StopWatching() {
	if w.isWatching.Load() {
		close(w.stopChan)
		w.wg.Wait()
	}
}

// IsWatching returns the current watching status
func (w *WatcherImpl[T]) IsWatching() bool {
	return w.isWatching.Load()
}

func (w *WatcherImpl[T]) watch() error {
	_ = w.storeImpl.ReconcileList()
	cur_res_list, _ := w.storeImpl.List()
	for _, res := range cur_res_list {
		id := res.GetID()
		//	log.Printf("Reconciling resource: %s\n", id)
		err := res.Reconcile()
		if err != nil {
			ErrorLogger.Printf("Error reconciling resource: %s, Error: %v\n", id, err)
		}
		status, _ := res.GetStatus()
		err = w.storeImpl.UpdateStatus(id, status)
		if err != nil {
			ErrorLogger.Printf("Error updating resource status: %s, Error: %v\n", id, err)
		}
		//InfoLogger.Printf("Resource status: %s\n", status)
	}
	return nil
}

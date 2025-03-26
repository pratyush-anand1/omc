package watcher

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/enrayga/omc-o2ims/internal/operator/resource"
	//"github.com/enrayga/omc-o2ims/internal/operator/store"
	"github.com/stretchr/testify/assert"
)

type MockStore[T resource.Resource] struct {
	list            []T
	reconcileErr    error
	updateStatusErr error
}

func NewMockStore[T resource.Resource]() (*MockStore[T], error) {
	return &MockStore[T]{}, nil
}

func (ms *MockStore[T]) List() ([]T, error) {
	return ms.list, nil
}

func (ms *MockStore[T]) ReconcileList() error {
	return ms.reconcileErr
}

func (ms *MockStore[T]) UpdateStatus(id string, status map[string]interface{}) error {
	return ms.updateStatusErr
}

type MockResource struct {
	id     string
	status map[string]interface{}
}

func (mr *MockResource) GetID() string {
	return mr.id
}

func (mr *MockResource) SetInitFields(name string, fields map[string]interface{}) error {
	return nil
}

func (mr *MockResource) Compare(name string, fields map[string]interface{}, apply bool) (bool, error) {
	return false, nil
}

func (mr *MockResource) GetNew() interface{} {
	return nil
}

func (mr *MockResource) GetName() string {
	return ""
}

func (mr *MockResource) GetType() (map[string]string, error) {
	return nil, nil
}

func (mr *MockResource) GetMetadata() (map[string]interface{}, error) {
	return nil, nil
}

func (mr *MockResource) GetSpec() (map[string]interface{}, error) {
	return nil, nil
}

func (mr *MockResource) GetStatus() (map[string]interface{}, error) {
	return mr.status, nil
}

func (mr *MockResource) Reconcile() error {
	return nil
}

func (mr *MockResource) SetDeleteFlag() error {
	return nil
}

func (mr *MockResource) GetDeleteFlag() bool {
	return false
}

var mu sync.Mutex

func TestWatcherStartStop(t *testing.T) {
	t.Run("Test Simple StartWatching and StopWatching", func(t *testing.T) {
		ctx := context.Background()
		store := &MockStore[*MockResource]{}

		watcher := NewWatcherImpl[*MockResource](store)
		err := watcher.StartWatching(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		assert.True(t, watcher.IsWatching())

		watcher.StopWatching()
		time.Sleep(100 * time.Millisecond)

		assert.False(t, watcher.IsWatching())
	})

	t.Run("Test StartWatching for already running watcher", func(t *testing.T) {
		ctx := context.Background()
		store := &MockStore[*MockResource]{}

		watcher := NewWatcherImpl[*MockResource](store)
		err := watcher.StartWatching(ctx)
		assert.NoError(t, err)

		err = watcher.StartWatching(ctx)
		assert.Error(t, err)

		watcher.StopWatching()
	})

	t.Run("Test StartWatching context cancelling should stop watching", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		store := &MockStore[*MockResource]{}

		watcher := NewWatcherImpl[*MockResource](store)
		cancel()

		err := watcher.StartWatching(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		assert.False(t, watcher.IsWatching())
	})
}

func TestWatcherAddAndDeleteResources(t *testing.T) {
	t.Run("Test case to verify the addition of a newly discovered resource", func(t *testing.T) {
		ctx := context.Background()
		store := &MockStore[*MockResource]{}
		watcher := NewWatcherImpl[*MockResource](store)

		mockResource := &MockResource{id: "resource-1", status: map[string]interface{}{"status": "ready"}}
		store.list = []*MockResource{mockResource}

		err := watcher.StartWatching(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		watcher.StopWatching()
	})
}

func TestWatcherUpdateLastInvocationTime(t *testing.T) {
	t.Run("Test UpdatelastInvocationTime and GetUpdatelastInvocationTime", func(t *testing.T) {
		store := &MockStore[*MockResource]{}
		watcher := NewWatcherImpl[*MockResource](store)

		watcher.UpdatelastInvocationTime()
		lastInvocationTime := watcher.GetUpdatelastInvocationTime()

		assert.False(t, lastInvocationTime.IsZero())
	})
}

func TestWatcherWatch(t *testing.T) {
	t.Run("Test watch method with reconcile error", func(t *testing.T) {
		ctx := context.Background()
		store := &MockStore[*MockResource]{reconcileErr: fmt.Errorf("reconcile error")}
		watcher := NewWatcherImpl[*MockResource](store)

		err := watcher.StartWatching(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		watcher.StopWatching()
	})

	t.Run("Test watch method with update status error", func(t *testing.T) {
		ctx := context.Background()
		mockResource := &MockResource{id: "resource-1", status: map[string]interface{}{"status": "ready"}}
		store := &MockStore[*MockResource]{list: []*MockResource{mockResource}, updateStatusErr: fmt.Errorf("update status error")}
		watcher := NewWatcherImpl[*MockResource](store)

		err := watcher.StartWatching(ctx)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		watcher.StopWatching()
	})
}

func TestWatcherInit(t *testing.T) {
	t.Run("Test Init method", func(t *testing.T) {
		store := &MockStore[*MockResource]{}
		watcher := &WatcherImpl[*MockResource]{}

		err := watcher.Init(store)
		assert.NoError(t, err)
		assert.Equal(t, store, watcher.storeImpl)
		assert.NotNil(t, watcher.stopChan)
	})
}

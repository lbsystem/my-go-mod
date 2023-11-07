package mystruct

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type MyTimeStrck[K comparable] struct {
	Key      K
	Deadline time.Time
}

type MyTimeList[K comparable] []MyTimeStrck[K]

func (h MyTimeList[K]) Len() int           { return len(h) }
func (h MyTimeList[K]) Less(i, j int) bool { return h[i].Deadline.Before(h[j].Deadline) }
func (h MyTimeList[K]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MyTimeList[K]) Push(x interface{}) {
	*h = append(*h, x.(MyTimeStrck[K]))
}

func (h *MyTimeList[K]) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type MyData[K comparable, V any] struct {
	data          sync.Map
	timeList      *MyTimeList[K]
	count         int64
	LimitSize     int
	checkInterval time.Duration
	mu            sync.Mutex
}

func NewLimterTime[K comparable, V any](ctx context.Context, size int, interval time.Duration) *MyData[K, V] {
	newData := &MyData[K, V]{}
	newData.LimitSize = size
	newData.checkInterval = interval
	newData.timeList = &MyTimeList[K]{}
	heap.Init(newData.timeList)
	go newData.checkExpire(ctx)
	return newData
}

func (d *MyData[K, V]) Get(key K) (value V, ok bool) {
	val, ok := d.data.Load(key)
	if !ok {
		return
	}
	return val.(V), ok
}

func (d *MyData[K, V]) Add(key K, value V, t time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock() // 使用 defer 确保在函数退出时释放锁

	// 检查键是否已存在，如果存在则直接返回
	_, exists := d.data.LoadOrStore(key, value)
	if exists {
		return
	}

	// 如果达到限制大小，则删除最早添加的元素
	if d.LimitSize != 0 && d.Count() >= d.LimitSize {
		a := heap.Pop(d.timeList)
		v, ok := a.(MyTimeStrck[K])
		if ok {
			d.data.Delete(v.Key)
			atomic.AddInt64(&d.count, -1)
		}
	}

	// 添加新元素到堆中
	item := MyTimeStrck[K]{Key: key, Deadline: time.Now().Add(t)}
	heap.Push(d.timeList, item)
	atomic.AddInt64(&d.count, 1)
}

func (d *MyData[K, V]) Del(key K) {
	d.mu.Lock()
	if _, loaded := d.data.LoadAndDelete(key); loaded {
		atomic.AddInt64(&d.count, -1)
		for i, item := range *d.timeList {
			if item.Key == key {
				heap.Remove(d.timeList, i)
				break
			}
		}
	}
	d.mu.Unlock()
}

func (d *MyData[K, V]) checkExpire(ctx context.Context) {
	ticker := time.NewTicker(d.checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.mu.Lock()
			now := time.Now()
			for d.timeList.Len() > 0 && (*d.timeList)[0].Deadline.Before(now) {

				key := (*d.timeList)[0].Key
				fmt.Println("del ", key)
				heap.Pop(d.timeList)
				d.data.Delete(key)
				atomic.AddInt64(&d.count, -1)
			}
			d.mu.Unlock()
		}
	}
}

func (d *MyData[K, V]) Count() int {
	return int(atomic.LoadInt64(&d.count))
}

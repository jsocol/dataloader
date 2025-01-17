package dataloader

import (
	"fmt"
	"sync"
	"time"
)

type Error[K comparable] interface {
	error
	Key() K
}

type keyError[K comparable] struct {
	msg string
	key K
}

func (k *keyError[K]) Error() string {
	return fmt.Sprintf("%s (%v)", k.msg, k.key)
}

func (k *keyError[K]) Key() K {
	return k.key
}

type result[T any] struct {
	value T
	err   error
}

// The Fetcher function should take a list of keys and return a map of keys to
// values. This may involve network requests or other slow or expensive calls.
type Fetcher[K comparable, V any] func([]K) (map[K]V, error)

type config struct {
	delay    time.Duration
	maxBatch int
}

// Loader is a generic implementation of the GraphQL "data loader" pattern that
// collapses several individual lookups by a key into one lookup as a list.
type Loader[K comparable, V any] struct {
	mu      sync.Mutex
	tasks   map[K][]chan *result[V]
	fetcher Fetcher[K, V]
	tick    <-chan time.Time
	config  config
}

func New[K comparable, V any](fetchFn Fetcher[K, V], opts ...Option) *Loader[K, V] {
	c := config{
		delay: time.Millisecond,
	}
	for _, o := range opts {
		o(&c)
	}
	return &Loader[K, V]{
		fetcher: fetchFn,
		tasks:   make(map[K][]chan *result[V]),
		config:  c,
	}
}

func (l *Loader[K, V]) Load(key K) (V, error) {
	ch := make(chan *result[V], 1)
	defer close(ch)

	l.enqueue(key, ch)
	res := <-ch
	return res.value, res.err
}

func (l *Loader[K, V]) LoadMany(keys ...K) ([]V, []error) {
	ret := make([]V, 0, len(keys))
	var errs []error
	var mu sync.Mutex

	var wg sync.WaitGroup
	wg.Add(len(keys))
	for i, k := range keys {
		go func(i int, k K) {
			defer wg.Done()
			ch := make(chan *result[V], 1)
			defer close(ch)

			l.enqueue(k, ch)
			res := <-ch

			mu.Lock()
			defer mu.Unlock()

			if res.err != nil {
				errs = append(errs, res.err)
				return
			}
			ret = append(ret, res.value)
		}(i, k)
	}
	wg.Wait()
	return ret, errs
}

func (l *Loader[K, V]) enqueue(k K, ch chan *result[V]) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tasks[k] = append(l.tasks[k], ch)

	if l.config.maxBatch > 0 && len(l.tasks) >= l.config.maxBatch {
		// if we've hit the max batch size, initiate a fetch immediately
		go func() {
			l.fetch()
		}()
	} else if l.tick == nil {
		// if we aren't waiting yet, start waiting
		l.tick = time.After(l.config.delay)
		go func() {
			<-l.tick
			l.fetch()
		}()
	}
}

func (l *Loader[K, V]) fetch() {
	l.mu.Lock()
	defer l.mu.Unlock()

	keys := make([]K, 0, len(l.tasks))
	for k := range l.tasks {
		keys = append(keys, k)
	}

	results, err := l.fetcher(keys)
	if err != nil {
		l.sendError(err)
		return
	}

	for k, v := range results {
		chans := l.tasks[k]
		if chans == nil {
			panic(fmt.Errorf("task key missing: %v", k))
		}
		res := &result[V]{
			value: v,
		}
		for _, ch := range chans {
			ch <- res
		}
		delete(l.tasks, k)
	}

	// handle the requests with no result
	if len(l.tasks) > 0 {
		for k, chans := range l.tasks {
			res := &result[V]{
				err: &keyError[K]{
					msg: "not found",
					key: k,
				},
			}
			for _, ch := range chans {
				ch <- res
			}
			delete(l.tasks, k)
		}
	}

	l.tick = nil
}

func (l *Loader[K, V]) sendError(err error) {
	res := &result[V]{
		err: err,
	}

	for _, chans := range l.tasks {
		for _, ch := range chans {
			ch <- res
		}
	}
}

package dataloader_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsocol/dataloader"
)

func TestLoad(t *testing.T) {
	var calls atomic.Int64
	fetcher := func(keys []string) (map[string]int, error) {
		calls.Add(1)
		assert.ElementsMatch(t, []string{"f", "ab", "ef"}, keys)
		ret := make(map[string]int, len(keys))
		for _, k := range keys {
			ret[k] = len(k)
		}
		return ret, nil
	}

	l := dataloader.New(fetcher)

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		v, err := l.Load("f")
		assert.NoError(t, err)
		assert.Equal(t, 1, v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load("ab")
		assert.NoError(t, err)
		assert.Equal(t, 2, v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load("ef")
		assert.NoError(t, err)
		assert.Equal(t, 2, v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load("f")
		assert.NoError(t, err)
		assert.Equal(t, 1, v)
	}()

	wg.Wait()

	assert.Equal(t, int64(1), calls.Load())
}

type key struct {
	major, minor string
}

func TestComplexKeys(t *testing.T) {
	var calls atomic.Int64
	fetcher := func(keys []key) (map[key]string, error) {
		assert.Len(t, keys, 1)
		calls.Add(1)

		ret := make(map[key]string, len(keys))
		for _, k := range keys {
			ret[k] = fmt.Sprintf("%s-%s", k.major, k.minor)
		}
		return ret, nil
	}

	l := dataloader.New(fetcher)

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		v, err := l.Load(key{major: "resource", minor: "nested"})
		assert.NoError(t, err)
		assert.Equal(t, "resource-nested", v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load(key{major: "resource", minor: "nested"})
		assert.NoError(t, err)
		assert.Equal(t, "resource-nested", v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load(key{major: "resource", minor: "nested"})
		assert.NoError(t, err)
		assert.Equal(t, "resource-nested", v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load(key{major: "resource", minor: "nested"})
		assert.NoError(t, err)
		assert.Equal(t, "resource-nested", v)
	}()

	wg.Wait()

	assert.Equal(t, int64(1), calls.Load())
}

func TestPartialNotFound(t *testing.T) {
	var calls atomic.Int64
	fetcher := func(keys []string) (map[string]string, error) {
		calls.Add(1)
		assert.Len(t, keys, 4)
		return map[string]string{
			"foo": "yes-foo",
			"bar": "yes-bar",
		}, nil
	}

	l := dataloader.New(fetcher)

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		v, err := l.Load("foo")
		assert.NoError(t, err)
		assert.Equal(t, "yes-foo", v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load("bar")
		assert.NoError(t, err)
		assert.Equal(t, "yes-bar", v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load("baz")

		assert.ErrorContains(t, err, "not found")
		assert.ErrorContains(t, err, "baz")
		kErr, ok := err.(dataloader.Error[string])
		assert.True(t, ok, "error should have type dataloader.Error[string]")
		assert.Equal(t, "baz", kErr.Key())

		assert.Empty(t, v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load("quux")
		assert.ErrorContains(t, err, "not found")
		assert.ErrorContains(t, err, "quux")
		assert.Empty(t, v)
	}()

	wg.Wait()

	assert.Equal(t, int64(1), calls.Load())
}

func TestMaxBatchSize(t *testing.T) {
	var calls atomic.Int64
	fetcher := func(keys []string) (map[string]int, error) {
		calls.Add(1)
		ret := make(map[string]int, len(keys))
		for _, k := range keys {
			ret[k] = len(k)
		}
		return ret, nil
	}

	l := dataloader.New(fetcher, dataloader.WithMaxBatch(2))

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		v, err := l.Load("f")
		assert.NoError(t, err)
		assert.Equal(t, 1, v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load("ab")
		assert.NoError(t, err)
		assert.Equal(t, 2, v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load("ef")
		assert.NoError(t, err)
		assert.Equal(t, 2, v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load("f")
		assert.NoError(t, err)
		assert.Equal(t, 1, v)
	}()

	wg.Wait()

	assert.Equal(t, int64(2), calls.Load())
}

func TestLoadMany(t *testing.T) {
	var calls atomic.Int64
	fetcher := func(keys []string) (map[string]string, error) {
		calls.Add(1)
		return map[string]string{
			"foo": "yes-foo",
			"bar": "yes-bar",
		}, nil
	}

	l := dataloader.New(fetcher)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		vs, errs := l.LoadMany("foo", "bar")
		assert.Empty(t, errs)
		assert.ElementsMatch(t, []string{"yes-foo", "yes-bar"}, vs)
	}()

	go func() {
		defer wg.Done()
		vs, errs := l.LoadMany("foo", "quux")

		assert.Len(t, errs, 1)
		assert.ErrorContains(t, errs[0], "quux")
		kErr, ok := errs[0].(dataloader.Error[string])
		assert.True(t, ok, "error should have type dataloader.Error[string]")
		assert.Equal(t, "quux", kErr.Key())

		assert.Len(t, vs, 1)
		assert.Equal(t, "yes-foo", vs[0])
	}()

	wg.Wait()

	assert.Equal(t, int64(1), calls.Load())
}

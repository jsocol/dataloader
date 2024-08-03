package restdataloader_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	restdataloader "github.com/jsocol/rest-data-loader"
)

func TestDebounce(t *testing.T) {
	var calls atomic.Int64
	fetcher := func(keys []string) (map[string]int, error) {
		calls.Add(1)
		assert.ElementsMatch(t, []string{"f", "ab", "ef"}, keys)

		time.Sleep(time.Millisecond)
		ret := make(map[string]int, len(keys))
		for _, k := range keys {
			ret[k] = len(k)
		}
		return ret, nil
	}

	l := restdataloader.New(fetcher)

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

	l := restdataloader.New(fetcher)

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

	l := restdataloader.New(fetcher)

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
		assert.ErrorIs(t, err, restdataloader.NotFound)
		assert.Empty(t, v)
	}()

	go func() {
		defer wg.Done()
		v, err := l.Load("quux")
		assert.ErrorIs(t, err, restdataloader.NotFound)
		assert.Empty(t, v)
	}()

	wg.Wait()

	assert.Equal(t, int64(1), calls.Load())
}

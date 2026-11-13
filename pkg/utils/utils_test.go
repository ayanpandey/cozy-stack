package utils

import (
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	s1 := RandomString(10)
	s2 := RandomString(20)

	assert.Len(t, s1, 10)
	assert.Len(t, s2, 20)
	assert.NotEqual(t, s1, s2)
}

func TestRandomStringFast(t *testing.T) {
	// Same seed produces reproducible results
	rng1 := rand.New(rand.NewSource(42))
	rng2 := rand.New(rand.NewSource(42))

	s1 := RandomStringFast(rng1, 10)
	s2 := RandomStringFast(rng1, 20)
	s3 := RandomStringFast(rng2, 10)
	s4 := RandomStringFast(rng2, 20)

	assert.Len(t, s1, 10)
	assert.Len(t, s2, 20)
	assert.Equal(t, s1, s3)
	assert.Equal(t, s2, s4)
}

func TestRandomStringConcurrentAccess(t *testing.T) {
	n := 10000
	var wg sync.WaitGroup
	wg.Add(n)

	ms := make(map[string]struct{})
	var mu sync.Mutex

	gotDup := false

	for i := 0; i < n; i++ {
		go func() {
			s := RandomString(10)
			defer wg.Done()
			mu.Lock()
			defer mu.Unlock()
			if _, ok := ms[s]; ok {
				gotDup = true
			}
			var q struct{}
			ms[s] = q
		}()
	}
	wg.Wait()

	if gotDup {
		t.Fatal("should be unique strings")
	}
}

func TestStripPort(t *testing.T) {
	d1 := StripPort("localhost")
	assert.Equal(t, "localhost", d1)
	d2 := StripPort("localhost:8080")
	assert.Equal(t, "localhost", d2)
	d3 := StripPort("localhost:8080:8081")
	assert.Equal(t, "localhost:8080:8081", d3)
}

func TestSplitTrimString(t *testing.T) {
	parts1 := SplitTrimString("", ",")
	assert.EqualValues(t, []string{}, parts1)
	parts2 := SplitTrimString("foo,bar,baz,", ",")
	assert.EqualValues(t, []string{"foo", "bar", "baz"}, parts2)
	parts3 := SplitTrimString(",,,,", ",")
	assert.EqualValues(t, []string{}, parts3)
	parts4 := SplitTrimString("foo  ,, bar,  baz  ,", ",")
	assert.EqualValues(t, []string{"foo", "bar", "baz"}, parts4)
	parts5 := SplitTrimString("    ", ",")
	assert.EqualValues(t, []string{}, parts5)
}

func TestFileExists(t *testing.T) {
	exists, err := FileExists("/no/such/file")
	assert.NoError(t, err)
	assert.False(t, exists)
	exists, err = FileExists("/etc/hosts")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestAbsPath(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NoError(t, err)
	assert.NotEmpty(t, home)
	assert.Equal(t, home, AbsPath("~"))
	foo := AbsPath("foo")
	wd, _ := os.Getwd()
	assert.Equal(t, wd+"/foo", foo)
	bar := AbsPath("~/bar")
	assert.Equal(t, home+"/bar", bar)
	baz := AbsPath("$HOME/baz")
	assert.Equal(t, home+"/baz", baz)
	qux := AbsPath("/qux")
	assert.Equal(t, "/qux", qux)
	quux := AbsPath("////qux//quux/../quux")
	assert.Equal(t, "/qux/quux", quux)
}

func TestParseBoundedInt(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		lo, hi    int
		def, want int
	}{
		{"empty", "", 0, 1000, 100, 100},
		{"not a number", "abc", 0, 1000, 100, 100},
		{"negative", "-1", 0, 1000, 100, 100},
		{"in range", "5", 0, 1000, 100, 5},
		{"zero", "0", 0, 1000, 100, 0},
		{"at max", "1000", 0, 1000, 100, 1000},
		{"above max", "1001", 0, 1000, 100, 100},
		{"beyond int64", "99999999999999999999", 0, math.MaxInt, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ParseBoundedInt(tc.in, tc.lo, tc.hi, tc.def))
		})
	}

	t.Run("does not wrap a value beyond int", func(t *testing.T) {
		got := ParseBoundedInt("4294967297", 0, math.MaxInt, 0)
		if strconv.IntSize == 32 {
			assert.Equal(t, 0, got)
		} else {
			assert.Equal(t, int64(4294967297), int64(got))
		}
	})
}

package log

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	K = 1 << 10
	M = 1 << 20
	G = 1 << 30
)

type RotateFile struct {
	mutex sync.Mutex

	outFh          *os.File
	curFn          string
	lastRotateTime time.Time
	rotateSize     int64
	rotateTime     time.Duration
	backTime       time.Duration
}

type Options struct {
	rotateTime time.Duration
	backTime   time.Duration

	rotateSize int64
}

func (o *Options) apply() {
	if o.backTime == 0 {
		o.backTime = time.Hour * 24 * 7
	}

	if o.rotateTime == 0 {
		o.rotateTime = time.Hour * 24
	}

	if o.rotateSize == 0 {
		o.rotateSize = 1 * G
	}
}

type RotateOption func(*Options)

func WithBackTime(b time.Duration) RotateOption {
	return func(o *Options) {
		o.backTime = b
	}
}

func WithRotateTime(r time.Duration) RotateOption {
	return func(o *Options) {
		o.rotateTime = r
	}
}

func WithRotateSize(s int64) RotateOption {
	return func(o *Options) {
		o.rotateSize = s
	}
}

func NewRotateFile(path string, opts ...RotateOption) *RotateFile {
	ops := Options{}

	for _, o := range opts {
		o(&ops)
	}
	ops.apply()

	r := &RotateFile{
		curFn:          path,
		rotateTime:     ops.rotateTime,
		backTime:       ops.backTime,
		rotateSize:     ops.rotateSize,
		lastRotateTime: time.Now(),
	}

	go r.loop()
	return r
}

func (r *RotateFile) Write(p []byte) (n int, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.write(p)
}

func (r *RotateFile) write(p []byte) (n int, err error) {
	err = r.write_nolock()
	if err != nil {
		return
	}

	return r.outFh.Write(p)
}

func (r *RotateFile) write_nolock() (err error) {
	if !fileExist(r.curFn) {
		return r.createFile()
	}

	now := time.Now()

	if now.Sub(r.lastRotateTime) < r.rotateTime &&
		isSameDay(now, r.lastRotateTime) &&
		r.rotateSize > fileSize(r.curFn) {
		return
	}

	r.lastRotateTime = now
	oldFn := genFileName(r.curFn, now)
	if err = os.Rename(r.curFn, oldFn); err != nil {
		return
	}

	return r.createFile()
}

func (r *RotateFile) createFile() (err error) {
	// if we got here, then we need to create a file
	fh, err := os.OpenFile(r.curFn, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}

	if r.outFh != nil {
		r.outFh.Close()
	}

	r.outFh = fh
	r.lastRotateTime = time.Now()
	return
}

func (r *RotateFile) loop() {
	tc := time.NewTicker(time.Minute)
	for {
		select {
		case <-tc.C:
			allFiles := getAllFiles(r.curFn)
			for _, b := range filterBackFiles(allFiles, r.curFn, r.backTime) {
				os.Remove(b)
			}
		}
	}
}

func getAllFiles(fn string) (files []string) {
	dirPath := filepath.Dir(fn)
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return
	}

	prefix := fn
	sep := string(os.PathSeparator)
	idx := strings.LastIndex(prefix, sep)
	if idx != -1 {
		prefix = prefix[idx+1:]
	}

	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}

		if !strings.HasPrefix(fi.Name(), prefix) {
			continue
		}

		files = append(files, fn[:idx+1]+fi.Name())
	}

	return
}

func filterBackFiles(files []string, curFn string, backTime time.Duration) (bf []string) {
	now := time.Now()
	backedTime := now.Add(-1 * backTime)

	for _, f := range files {
		if f != curFn && f < genFileName(curFn, backedTime) {
			bf = append(bf, f)
		}
	}

	return
}

func genFileName(curFn string, now time.Time) string {
	return fmt.Sprintf("%v.%v.%v", curFn, now.Format("2006-01-02"), now.Format("150405"))
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	if err != nil {
		panic(err)
	}

	return true
}

func isSameDay(l time.Time, r time.Time) bool {
	return l.Format("2006-01-02") == r.Format("2006-01-02")
}

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return math.MinInt64
	}

	return fi.Size()
}

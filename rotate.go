package rotate

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	leftSymbol  byte        = '['
	rightSymbol byte        = ']'
	perm        os.FileMode = 0644
)

const (
	_  = iota
	KB = 1 << (10 * iota)
	MB = 1 << (10 * iota)
	GB = 1 << (10 * iota)
)

const (
	Day   = time.Hour * 24
	Week  = Day * 7
	Month = Day * 30
)

// SetSymbol Set the left and right symbols of the file slice sequence number.
func SetSymbol(left, right byte) {
	if left > 0 && right > 0 {
		leftSymbol = left
		rightSymbol = right
	}
}

// SetPerm Set permissions for creating files and folders.
func SetPerm(permissions os.FileMode) {
	perm = permissions
}

// ExpFunc Procedure for handler expired files.
type ExpFunc func([]string)

// Rotate The structure of the rotation file.
type Rotate struct {
	dir    string // storage folder path
	layout string // generate filename format template
	ext    string // file extension

	rotateDur           time.Duration  // rotation interval
	loc                 *time.Location // time zone settings
	maxAge              time.Duration  // maximum survival time
	maxSize             int64          // file size limit
	expiredFunc         ExpFunc        // processing of expired files
	deleteEmptyFileFlag bool           // remove empty file flag
	deleteEmptyDirFlag  bool           // remove empty dir flag

	mutex     *sync.RWMutex // lock
	index     int64         // rotated file index
	writeSize int64         // the size of the file content that has been written
	close     chan struct{} // close goroutine channel
	file      *os.File      // current file
	tick      *time.Timer   // timing rotation timer
}

// New create Rotate, Requires directory to store logs, time formatted layout, file extension.
func New(dir, layout, ext string, options ...Option) (*Rotate, error) {
	if len(dir) == 0 || len(layout) == 0 || len(ext) == 0 {
		return nil, fmt.Errorf("dir, layout, ext name len equel zero")
	}
	if !strings.HasPrefix(dir, "/") {
		if dir[1] != ':' && !strings.HasPrefix(dir, "./") {
			dir = fmt.Sprintf("./%s", dir)
		}
	}
	err := os.MkdirAll(dir, perm)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(ext, ".") {
		ext = fmt.Sprintf(".%s", ext)
	}
	r := &Rotate{
		// required parameters
		dir:    dir,
		layout: layout,
		ext:    ext,

		// optional parameters (defaults value)
		rotateDur:           Day,
		loc:                 time.Local,
		maxAge:              0,
		maxSize:             0,
		expiredFunc:         nil,
		deleteEmptyFileFlag: true,
		deleteEmptyDirFlag:  true,

		// private parameters
		mutex:     new(sync.RWMutex),
		index:     1,
		writeSize: 0,
		close:     make(chan struct{}),
	}
	for _, option := range options {
		option(r)
	}
	now := nowRotateTime(r.rotateDur, r.loc)
	ok, _, index := r.getNowFilenameMax(now)
	if ok {
		index++
	} else {
		index = 0
	}
	file, err := r.open(r.genFilename(now, index))
	if err != nil {
		return nil, err
	}
	r.file = file
	r.tick = time.NewTimer(calcDur(time.Now(), now.Add(r.rotateDur)))
	go r.task()
	r.expiredFileHandler()
	return r, nil
}

// implement the io.Writer interface
func (r *Rotate) Write(data []byte) (int, error) {
	var (
		n   int
		err error
	)
	r.mutex.RLock()
	if r.file == nil {
		err = fmt.Errorf("file is close")
	} else {
		n, err = r.file.Write(data)
	}
	r.mutex.RUnlock()

	if err != nil {
		return n, err
	}

	r.mutex.Lock()
	r.writeSize += int64(n)
	r.mutex.Unlock()

	if r.maxSize > 0 && r.writeSize >= r.maxSize {
		r.rotateSize()
	}
	return n, err
}

// Sync method
func (r *Rotate) Sync() error {
	return r.file.Sync()
}

// implement the io.Closer interface
func (r *Rotate) Close() error {
	r.close <- struct{}{}
	var f *os.File
	r.mutex.Lock()
	f = r.file
	r.file = nil
	r.mutex.Unlock()
	_ = f.Sync()
	return f.Close()
}

// goroutine tasks
func (r *Rotate) task() {
	for {
		select {
		case <-r.close:
			return
		case now := <-r.tick.C:
			r.tick.Reset(calcDur(now, nowRotateTime(r.rotateDur, r.loc).Add(r.rotateDur)))
			r.rotateTime(now)
			r.expiredFileHandler()
		}
	}
}

// rotateSize split by size
func (r *Rotate) rotateSize() {
	i := r.index + 1
	newFilename := r.genFilename(nowRotateTime(r.rotateDur, r.loc), i)
	f, err := r.open(newFilename)
	if err != nil {
		panic(err)
	}
	var oldFile *os.File

	r.mutex.Lock()
	oldFile = r.file
	r.file = f
	r.writeSize = 0
	r.index = i
	r.mutex.Unlock()

	if (i - 1) <= 1 {
		newFilename = r.genFilename(nowRotateTime(r.rotateDur, r.loc), 1)
		oldFilename := oldFile.Name()
		_ = oldFile.Sync()
		_ = oldFile.Close()
		_ = os.Rename(oldFilename, newFilename)
	} else {
		_ = oldFile.Sync()
		_ = oldFile.Close()
	}
}

// rotateTime split by time
func (r *Rotate) rotateTime(now time.Time) {
	filename := r.genFilename(now, 0)
	f, err := r.open(filename)
	if err != nil {
		panic(err)
	}

	var oldFile *os.File

	r.mutex.Lock()
	oldFile = r.file
	r.file = f
	r.writeSize = 0
	r.index = 1
	r.mutex.Unlock()

	if r.deleteEmptyFileFlag {
		filename = oldFile.Name()
		info, err := oldFile.Stat()
		_ = oldFile.Sync()
		_ = oldFile.Close()
		if err != nil {
			return
		}
		if info.Size() > 0 {
			return
		}
		_ = os.Remove(filename)
	} else {
		_ = oldFile.Sync()
		_ = oldFile.Close()
	}

	if r.deleteEmptyDirFlag {
		emptyDirs := getEmptyDirs(r.dir)
		for _, dir := range emptyDirs {
			_ = os.Remove(dir)
		}
	}
}

// open file by file name
func (r *Rotate) open(filename string) (*os.File, error) {
	dir, _ := path.Split(filename)
	err := os.MkdirAll(dir, perm)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, perm)
}

// genFilename generate file name
func (r *Rotate) genFilename(t time.Time, i int64) string {
	var buf strings.Builder
	buf.WriteString(join(r.dir, t.Format(r.layout)))
	if i >= 1 {
		buf.WriteByte(leftSymbol)
		buf.WriteString(fmt.Sprintf("%d", i))
		buf.WriteByte(rightSymbol)
	}
	buf.WriteString(r.ext)
	return buf.String()
}

// parseFilename To parse the file name, the suffix must be removed
func (r *Rotate) parseFilename(filename string) (bool, time.Time, int64) {
	rightIndex := len(filename) - 1
	leftIndex := 0
	if filename[rightIndex] == rightSymbol {
		leftIndex = rightIndex
		for 0 <= leftIndex {
			if filename[leftIndex] == leftSymbol {
				break
			}
			leftIndex--
		}
		if leftIndex < 0 {
			return false, time.Time{}, 0
		}
		i, err := strconv.ParseInt(filename[leftIndex+1:rightIndex], 10, 64)
		if err != nil {
			return false, time.Time{}, 0
		}
		t, err := time.ParseInLocation(join(r.dir, r.layout), filename[:leftIndex], r.loc)
		if err != nil {
			return false, time.Time{}, 0
		}
		return true, t, i
	} else {
		t, err := time.ParseInLocation(join(r.dir, r.layout), filename, r.loc)
		if err != nil {
			return false, time.Time{}, 0
		}
		return true, t, 0
	}
}

// getNowFilenameMax get the max chunk at this time
func (r *Rotate) getNowFilenameMax(now time.Time) (bool, string, int64) {
	var result string
	var fileOk bool
	var maxIndex int64 = 0
	files := getDirAllFiles(r.dir)
	if len(files) <= 0 {
		return fileOk, result, 0
	}
	for i := 0; i < len(files); i++ {
		ok, t, index := r.parseFilename(strings.TrimSuffix(files[i], r.ext))
		if ok && now.Equal(t) && index >= maxIndex {
			fileOk = true
			maxIndex = index
			result = files[i]
		}
	}
	return fileOk, result, maxIndex
}

// handle expired files
func (r *Rotate) expiredFileHandler() {
	if r.expiredFunc != nil && r.maxAge > 0 {
		exp := nowRotateTime(r.rotateDur, r.loc).Add(-r.maxAge)
		result := r.getExpirationFiles(exp)
		// prevent panic
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("expired handler panic:%v", err)
			}
		}()
		r.expiredFunc(result)
	}
}

// get expired files
func (r *Rotate) getExpirationFiles(exp time.Time) []string {
	var result []string
	files := getDirAllFiles(r.dir)
	for i := 0; i < len(files); i++ {
		ok, t, _ := r.parseFilename(strings.TrimSuffix(files[i], r.ext))
		if ok && (exp.Equal(t) || exp.After(t)) {
			result = append(result, files[i])
		}
	}
	return result
}

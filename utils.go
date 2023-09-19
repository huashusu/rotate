package rotate

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

// getNowRotateTime Used to obtain the naming format for time slices within a day.
// Preferably divisible by 24 Hour.
func nowRotateTime(dur time.Duration, loc *time.Location) time.Time {
	var base time.Time
	now := time.Now().In(loc)
	if loc != time.UTC {
		base = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.UTC)
		base = base.Truncate(dur)
		base = time.Date(base.Year(), base.Month(), base.Day(), base.Hour(), base.Minute(), base.Second(), base.Nanosecond(), loc)
	} else {
		base = now.Truncate(dur)
	}
	return base
}

// getDirAllFiles get all files in a directory, when an error occurs panic.
func getDirAllFiles(dir string) []string {
	var result []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		filename := join(dir, entry.Name())
		if entry.IsDir() {
			result = append(result, getDirAllFiles(filename)...)
		} else {
			result = append(result, filename)
		}
	}
	return result
}

// getEmptyDirs get all empty folders in a directory.
func getEmptyDirs(dir string) []string {
	var result []string
	dirs, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, entry := range dirs {
		if entry.IsDir() {
			name := join(dir, entry.Name())
			files, _ := os.ReadDir(name)
			if len(files) > 0 {
				entryDirs := getEmptyDirs(name)
				result = append(result, entryDirs...)
				var count int
				for _, file := range files {
					if file.IsDir() && hasString(entryDirs, join(name, file.Name())) {
						count++
					} else {
						break
					}
				}
				if count == len(files) {
					result = append(result, name)
				}
			} else {
				result = append(result, name)
			}
		}
	}
	return result
}

// hasString check if target exists in slice
func hasString(slice []string, target string) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == target {
			return true
		}
	}
	return false
}

// join path, if it is not an absolute path add './'.
func join(elem ...string) string {
	p := path.Join(elem...)
	if len(elem) > 0 && strings.HasPrefix(elem[0], "./") {
		p = fmt.Sprintf("./%s", p)
	}
	return p
}

// calcDur calculate time interval
func calcDur(start, end time.Time) time.Duration {
	return end.Sub(start)
}

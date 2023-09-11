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

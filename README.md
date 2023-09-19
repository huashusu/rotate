# Rotate

[![Golang](https://img.shields.io/badge/Golang-1.18-blue)](https://go.dev/)
[![GitHub](https://img.shields.io/github/license/huashusu/rotate)](https://github.com/Huashusu/rotate)

[English](README.md) | [简体中文](README.zh_CN.md)

## Preconditions

- **[Go](https://go.dev/)** version >= 1.18

> If you have tested it and can run it under a lower version, you can raise an `issue` or `pull requests` to change the
> docs and go.mod.

## Getting rotate

```shell
go get -u github.com/huashusu/rotate
```

## Documentation

[Document](https://godoc.org/github.com/Huashusu/rotate)

## Change log

[Change log](CHANGELOG.md)

## Example

```go
package main

import (
	"log"
	"os"
	"github.com/huashusu/rotate"
	"time"
)

func main() {
	file, err := rotate.New("log", "2006-01-02/15-04-05", "log",
		rotate.WithRotationDuration(time.Minute*1),
		rotate.WithDeleteEmptyFile(true),
		rotate.WithMaxSize(rotate.MB*10),
		rotate.WithMaxAge(time.Minute*5),
		rotate.WithExpiredHandler(func(expFiles []string) {
			for i := 0; i < len(expFiles); i++ {
				os.Remove(expFiles[i])
			}
		}),
	)
	if err != nil {
		panic(err)
	}
	log.SetOutput(file)
}
```

`New` method signature: `func New(dir, layout, ext string, options ...Option) (*Rotate, error)`

Parameter Description:

- `dir`: The root directory where log files are stored.
- `layout`: For the time template used in the generated log, the syntax used is the `format` method of the
  built-in `time` package, and there should be no illegal characters in the file name. The example is as follows:
    - No directory hierarchy exists: layout: `2006-01-02-15-04-05`
    - Separate logs by month: layout: `2006-01-02/15-04-05`
    - Sort logs by year, month and day: layout: `2006/01/02/15-04-05`
- `ext`: Log file suffix

## Option Parameters

### Set split time(`WithRotationDuration`)`default: Day`

**explain:** A truncating design was used for time selection. If the setting is less than one day, it is best to be
divisible by 24 (hours). The case is as follows:

current time: `2023/09/12 15:45:40`

- Set for 10 minutes: `2023/09/12 15:40:00`
- Set for 30 minutes: `2023/09/12 15:30:00`
- Set 1 hour: `2023/09/12 15:00:00`
- Set 6 hours: `2023/09/12 12:00:00`
- Set 1 day: `2023/09/12 00:00:00`

### Set time zone(`WithTimeZone`)`default: time.Local`

**explain:** The time template does not need to carry the time zone flag, and is set separately through the time zone,
which is used to truncate the time and parse the time from the file name.

### Set expiration time(`WithMaxAge`)`default: 0`

**explain:** Every time a file is split, it will be checked to see if there are any expired files.

### Set file size limit(`WithMaxSize`)`default: 0`

**explain:** After writing exceeds the file size, it will be split into log files with serial numbers.

### Set expiration handler(`WithExpiredHandler`)`default: nil`

**explain:** Usually used together with the expiration time, the path of the expired file is passed in, for example:
compressing the expired file, reading the file and sending it, deleting the file, etc.

### Set delete empty files(`WithDeleteEmptyFile`)`default: true`

**explain:** After the log is split, if the last file size is 0, delete it.

### Set delete empty directory(`WithDeleteEmptyDir`)`default: true`

**explain:** After deleting empty file logs, there may be empty folders. You can use this method to clean up empty
folders.

## Const

### Used for file size limits

`KB`: `1024`

`MB`: `1048576`

`GB`: `1073741824`

### Used for expiration time settings

`Day`: `time.Hour * 24`

`Week`: `Day * 7`

`Month`: `Day * 30`

## Static Method

`SetSymbol`: Set the left and right wrapping symbols of the serial number in the file, default
value:`Left`=`[` `Right`=`]`

`SetPerm`: Set the permission mode for creating folders and files, default value: `0644`
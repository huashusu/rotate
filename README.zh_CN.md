# Rotate

[![Golang](https://img.shields.io/badge/Golang-1.18-blue)](https://go.dev/)
[![GitHub](https://img.shields.io/github/license/huashusu/rotate)](https://github.com/Huashusu/rotate)

[English](README.md) | [简体中文](README.zh_CN.md)

## 前置条件

- **[Go](https://go.dev/)** version >= 1.18
> 如果您经过测试，可以在较低版本下运行，可提 `issue` 或 `pull requests` 变更版本前置和go.mod。

## 安装

```shell
go get -u github.com/huashusu/rotate
```

## 文档

[文档地址](https://godoc.org/github.com/Huashusu/rotate)

## 更新日志

[更新日志](CHANGELOG.zh_CN.md)

## 示例

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

`New`方法签名: `func New(dir, layout, ext string, options ...Option) (*Rotate, error)`

参数说明:

- `dir`: 存放日志文件的根目录
- `layout`: 对生成的日志的使用的时间模板，使用的语法是内置`time`包的`format`方法，其中不要存在文件名非法字符，示例如下：
    - 不存在目录层级：layout: `2006-01-02-15-04-05`
    - 按月份区分日志：layout: `2006-01-02/15-04-05`
    - 按年月日区分日志：layout: `2006/01/02/15-04-05`
- `ext`: 日志文件的后缀名

## 选项参数

### 设置切分时间(`WithRotationDuration`)`默认值: 一天`

**说明:** 时间选择上使用了取整设计。如果设置小于一天的最好能被24(小时)整除，案例如下：

当前时间: `2023/09/12 15:45:40`

- 设置10分钟: `2023/09/12 15:40:00`
- 设置30分钟: `2023/09/12 15:30:00`
- 设置1小时: `2023/09/12 15:00:00`
- 设置6小时: `2023/09/12 12:00:00`
- 设置1天: `2023/09/12 00:00:00`

### 设置时区(`WithTimeZone`)`默认值: time.Local`

**说明:** 时间模板上就可以不携带时区标志，通过时区另外设置，用于取整时间、从文件名解析时间

### 设置过期时间(`WithMaxAge`)`默认值: 0`

**说明:** 每次切分文件后，都会检查有没有过期文件

### 设置文件大小限制(`WithMaxSize`)`默认值: 0`

**说明:** 写入超过文件大小后，就会切分，带有序号的日志文件

### 设置过期处理程序(`WithExpiredHandler`)`默认值: nil`

**说明:** 通常与过期时间一起使用，传入的是过期文件的路径，例如：压缩过期文件、读取文件并发送、删除文件等等

### 设置删除空文件(`WithDeleteEmptyFile`)`默认值: true`

**说明:** 日志切分之后，如果上次文件大小为0，则删除

### 设置删除空文件夹(`WithDeleteEmptyDir`)`默认值: true`

**说明:** 删除空文件日志之后，有可能会存在空文件夹，可以使用这个方法清理空文件夹

## 常量

### 用于文件大小限制

`KB`: `1024`

`MB`: `1048576`

`GB`: `1073741824`

### 用于过期时间设置

`Day`: `time.Hour * 24`

`Week`: `Day * 7`

`Month`: `Day * 30`

## 静态方法

`SetSymbol`: 设置文件中存在序号的左、右包裹符号，默认值:`Left`=`[` `Right`=`]`

`SetPerm`: 设置创建文件夹、文件的权限模式，默认值: `0644`
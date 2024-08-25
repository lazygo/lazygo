package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// fileLogWriter implements LogWriter Interface.
type fileLogWriter struct {
	sync.RWMutex // write log order by order and  atomic incr maxLinesCurLines and maxSizeCurSize

	// The opened file
	Filename string `json:"filename"`
	writer   *os.File

	// Rotate at line
	MaxLines         int `json:"max_lines"`
	maxLinesCurLines int

	MaxFiles         int `json:"max_files"`
	MaxFilesCurFiles int

	// Rotate at size
	MaxSize        int `json:"maxsize"`
	maxSizeCurSize int

	// Rotate daily
	Daily         bool  `json:"daily"`
	MaxDays       int64 `json:"max_days"`
	dailyOpenDate int
	dailyOpenTime time.Time

	// Rotate hourly
	Hourly         bool  `json:"hourly"`
	MaxHours       int64 `json:"max_hours"`
	hourlyOpenDate int
	hourlyOpenTime time.Time

	Rotate bool `json:"rotate"`

	Perm string `json:"perm"`

	DirPerm string `json:"dir_perm"`

	RotatePerm string `json:"rotate_perm"`

	fileNameOnly, suffix string // like "project.log", project is fileNameOnly and .log is suffix
}

// newFileLogWriter create a FileLogWriter returning as LoggerInterface.
func newFileLogWriter(opt map[string]string) (logWriter, error) {
	fl := &fileLogWriter{
		Filename:   opt["filename"],
		Daily:      toBool(opt["daily"], false),
		MaxDays:    toInt64(opt["max_days"], 7),
		Hourly:     toBool(opt["hourly"], true),
		MaxHours:   toInt64(opt["max_hours"], 24*7),
		MaxLines:   toInt(opt["max_lines"], 0),
		MaxFiles:   toInt(opt["max_files"], 999),
		MaxSize:    toInt(opt["max_size"], 0),
		Perm:       "0777",
		DirPerm:    "0777",
		RotatePerm: "0440",
	}
	if opt["perm"] != "" {
		fl.Perm = opt["perm"]
	}
	if opt["rotate_perm"] != "" {
		fl.RotatePerm = opt["rotate_perm"]
	}

	if fl.Filename == "" {
		return nil, ErrInvalidFilename
	}
	fl.suffix = filepath.Ext(fl.Filename)
	fl.fileNameOnly = strings.TrimSuffix(fl.Filename, fl.suffix)
	if fl.suffix == "" {
		fl.suffix = ".log"
	}
	err := fl.startLogger()
	return fl, err
}

// start file logger. create log file and set to locker-inside file writer.
func (fl *fileLogWriter) startLogger() error {
	file, err := fl.createLogFile()
	if err != nil {
		return err
	}
	if fl.writer != nil {
		fl.writer.Close()
	}
	fl.writer = file
	return fl.initFd()
}

// Write write logger message into file.
func (fl *fileLogWriter) Write(b []byte, t time.Time, prefix string) (int, error) {
	hd, d, h := formatTimeHeader(t)
	fl.RLock()
	if fl.needRotateHourly(h) {
		fl.RUnlock()
		fl.Lock()
		if fl.needRotateHourly(h) {
			if err := fl.doRotate(t); err != nil {
				fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", fl.Filename, err)
			}
		}
		fl.Unlock()
	} else if fl.needRotateDaily(d) {
		fl.RUnlock()
		fl.Lock()
		if fl.needRotateDaily(d) {
			if err := fl.doRotate(t); err != nil {
				fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", fl.Filename, err)
			}
		}
		fl.Unlock()
	} else {
		fl.RUnlock()
	}

	fl.Lock()
	b = append(hd, b...)
	if prefix != "" {
		b = append([]byte(prefix+" "), b...)
	}
	n, err := fl.writer.Write(b)
	if err == nil {
		fl.maxLinesCurLines++
		fl.maxSizeCurSize += n
	}
	fl.Unlock()
	return n, err
}

// Close 关闭文件
func (fl *fileLogWriter) Close() error {
	fl.Flush()
	return fl.writer.Close()
}

// Flush flush file logger.
// there are no buffering messages in file logger in memory.
// flush file means sync file from disk.
func (fl *fileLogWriter) Flush() error {
	return fl.writer.Sync()
}

func (fl *fileLogWriter) needRotateDaily(day int) bool {
	return (fl.MaxLines > 0 && fl.maxLinesCurLines >= fl.MaxLines) ||
		(fl.MaxSize > 0 && fl.maxSizeCurSize >= fl.MaxSize) ||
		(fl.Daily && day != fl.dailyOpenDate)
}

func (fl *fileLogWriter) needRotateHourly(hour int) bool {
	return (fl.MaxLines > 0 && fl.maxLinesCurLines >= fl.MaxLines) ||
		(fl.MaxSize > 0 && fl.maxSizeCurSize >= fl.MaxSize) ||
		(fl.Hourly && hour != fl.hourlyOpenDate)

}

// createLogFile 创建日志文件
func (fl *fileLogWriter) createLogFile() (*os.File, error) {
	// Open the log file
	perm, err := strconv.ParseInt(fl.Perm, 8, 64)
	if err != nil {
		return nil, err
	}
	dirPerm, err := strconv.ParseInt(fl.DirPerm, 8, 64)
	if err != nil {
		return nil, err
	}

	dir := path.Dir(fl.Filename)
	err = os.MkdirAll(dir, os.FileMode(dirPerm))
	if err != nil {
		return nil, err
	}

	fd, err := os.OpenFile(fl.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
	if err != nil {
		return nil, err
	}

	// chmod
	err = os.Chmod(fl.Filename, os.FileMode(perm))
	return fd, err
}

func (fl *fileLogWriter) initFd() error {
	f := fl.writer
	fInfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("get stat err: %s", err)
	}
	fl.maxSizeCurSize = int(fInfo.Size())
	fl.dailyOpenTime = time.Now()
	fl.dailyOpenDate = fl.dailyOpenTime.Day()
	fl.hourlyOpenTime = time.Now()
	fl.hourlyOpenDate = fl.hourlyOpenTime.Hour()
	fl.maxLinesCurLines = 0
	if fl.Hourly {
		go fl.hourlyRotate(fl.hourlyOpenTime)
	} else if fl.Daily {
		go fl.dailyRotate(fl.dailyOpenTime)
	}
	if fInfo.Size() > 0 && fl.MaxLines > 0 {
		count, err := fl.lines()
		if err != nil {
			return err
		}
		fl.maxLinesCurLines = count
	}
	return nil
}

func (fl *fileLogWriter) dailyRotate(openTime time.Time) {
	y, m, d := openTime.Add(24 * time.Hour).Date()
	nextDay := time.Date(y, m, d, 0, 0, 0, 0, openTime.Location())
	tm := time.NewTimer(time.Duration(nextDay.UnixNano() - openTime.UnixNano() + 100))
	<-tm.C
	fl.Lock()
	if fl.needRotateDaily(time.Now().Day()) {
		if err := fl.doRotate(time.Now()); err != nil {
			fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", fl.Filename, err)
		}
	}
	fl.Unlock()
}

func (fl *fileLogWriter) hourlyRotate(openTime time.Time) {
	nextTime := openTime.Add(1 * time.Hour)
	y, m, d := nextTime.Date()
	h, _, _ := nextTime.Clock()
	nextHour := time.Date(y, m, d, h, 0, 0, 0, openTime.Location())
	tm := time.NewTimer(time.Duration(nextHour.UnixNano() - openTime.UnixNano() + 100))
	<-tm.C
	fl.Lock()
	if fl.needRotateHourly(time.Now().Hour()) {
		if err := fl.doRotate(time.Now()); err != nil {
			fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", fl.Filename, err)
		}
	}
	fl.Unlock()
}

func (fl *fileLogWriter) lines() (int, error) {
	fd, err := os.Open(fl.Filename)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	buf := make([]byte, 32768) // 32k
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := fd.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}

// DoRotate means it needs to write logs into a new file.
// new file name like xx.2013-01-01.log (daily) or xx.001.log (by line or size)
func (fl *fileLogWriter) doRotate(logTime time.Time) error {
	// file exists
	// Find the next available number
	num := fl.MaxFilesCurFiles + 1
	fName := ""
	format := ""
	var openTime time.Time
	rotatePerm, err := strconv.ParseInt(fl.RotatePerm, 8, 64)
	if err != nil {
		return err
	}

	_, err = os.Lstat(fl.Filename)
	if err != nil {
		// even if the file is not exist or other ,we should RESTART the logger
		goto RESTART
	}

	if fl.Hourly {
		format = "2006010215"
		openTime = fl.hourlyOpenTime
	} else if fl.Daily {
		format = "2006-01-02"
		openTime = fl.dailyOpenTime
	}

	// only when one of them be setted, then the file would be splited
	if fl.MaxLines > 0 || fl.MaxSize > 0 {
		for ; err == nil && num <= fl.MaxFiles; num++ {
			fName = fl.fileNameOnly + fmt.Sprintf(".%s.%03d%s", logTime.Format(format), num, fl.suffix)
			_, err = os.Lstat(fName)
		}
	} else {
		fName = fl.fileNameOnly + fmt.Sprintf(".%s%s", openTime.Format(format), fl.suffix)
		_, err = os.Lstat(fName)
		fl.MaxFilesCurFiles = num
	}

	// return error if the last file checked still existed
	if err == nil {
		return fmt.Errorf("Rotate: Cannot find free log number to rename %s", fl.Filename)
	}

	// close fileWriter before rename
	fl.writer.Close()

	// Rename the file to its new found name
	// even if occurs error,we MUST guarantee to  restart new logger
	err = os.Rename(fl.Filename, fName)
	if err != nil {
		goto RESTART
	}

	err = os.Chmod(fName, os.FileMode(rotatePerm))

RESTART:

	startLoggerErr := fl.startLogger()
	go fl.deleteOldLog()

	if startLoggerErr != nil {
		return fmt.Errorf("Rotate StartLogger: %s", startLoggerErr)
	}
	if err != nil {
		return fmt.Errorf("Rotate: %s", err)
	}
	return nil
}

// deleteOldLog 清理旧日志
func (fl *fileLogWriter) deleteOldLog() {
	dir := filepath.Dir(fl.Filename)
	absolutePath, err := filepath.EvalSymlinks(fl.Filename)
	if err == nil {
		dir = filepath.Dir(absolutePath)
	}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "Unable to delete old log '%s', error: %v\n", path, r)
			}
		}()

		if info == nil {
			return
		}
		if fl.Hourly {
			if !info.IsDir() && info.ModTime().Add(1*time.Hour*time.Duration(fl.MaxHours)).Before(time.Now()) {
				if strings.HasPrefix(filepath.Base(path), filepath.Base(fl.fileNameOnly)) &&
					strings.HasSuffix(filepath.Base(path), fl.suffix) {
					os.Remove(path)
				}
			}
		} else if fl.Daily {
			if !info.IsDir() && info.ModTime().Add(24*time.Hour*time.Duration(fl.MaxDays)).Before(time.Now()) {
				if strings.HasPrefix(filepath.Base(path), filepath.Base(fl.fileNameOnly)) &&
					strings.HasSuffix(filepath.Base(path), fl.suffix) {
					os.Remove(path)
				}
			}
		}
		return
	})
}

func init() {
	// 注册适配器
	registry.Add("file", newFileLogWriter)
}

package logger

import (
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

var (
	InfoLog  *log.Logger
	ErrorLog *log.Logger
)

func InitLogger() {
	errFile, err := os.OpenFile("errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	infoFile, err := os.OpenFile("info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	InfoLog = log.New(io.MultiWriter(os.Stderr, infoFile), "Info:", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	ErrorLog = log.New(io.MultiWriter(os.Stderr, errFile), "Error:", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		InfoLog.Printf("Request: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
		InfoLog.Printf("Response: %d", c.Writer.Status())
	}
}

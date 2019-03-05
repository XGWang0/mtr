package commlib

import (
	"io"
	"log"
	"os"
	"time"
)

var (
	Mtrloggger *log.Logger
	logerfile  = "/tmp/mtr/mtr_" + time.Now().Format("2006_01_02_15_04_08") + ".log"
)

func InitLogger() (*log.Logger, error) {
	filep, err := CreateFile(logerfile)
	if err != nil {
		return log.New(os.Stdout, "\r\n", log.LstdFlags|log.Lshortfile), err
	}
	mult_write := io.MultiWriter(filep, os.Stdout)
	return log.New(mult_write, "\r\n", log.LstdFlags|log.Lshortfile), nil
}

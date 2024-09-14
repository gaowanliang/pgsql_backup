package main

import (
	"io"
	"log"
	"os"
)

func setupLogging() (*os.File, error) {
	logFilePath := "backup.log"
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("无法创建日志文件: %v", err)
	}

	// 创建 MultiWriter，日志输出到文件和控制台
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return logFile, nil
}

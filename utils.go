package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

// createZipArchive creates a zip file with the specified WAL files
func createZipArchive(walFiles []string, zipFilePath string) error {
	// Create the zip file
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		log.Fatalf("无法创建 zip 文件: %v", err)
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add WAL files to the zip archive
	for _, file := range walFiles {
		err := addFileToZip(zipWriter, file)
		if err != nil {
			log.Fatalf("无法添加文件 %s 到 zip: %v", file, err)
		}
	}
	return nil
}

// addFileToZip adds a file to a zip archive
func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get the file name and create a header in the zip archive
	info, err := file.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate

	// Create a writer for the file
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// Copy the file content into the zip archive
	_, err = io.Copy(writer, file)
	return err
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

// executeCommand executes shell commands and returns the output or an error
func executeCommand(command string, args ...string) error {
	fmt.Println("执行命令:", command, args)
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// isFileModifyTimeOver checks if the file's modification time is over the specified duration
func isFileModifyTimeOver(filePath string, duration int) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Fatalf("无法获取文件信息: %v", err)
	}
	return fileInfo.ModTime().Before(time.Now().Add(-time.Duration(duration) * time.Hour))
}

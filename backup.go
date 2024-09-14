package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

// performFullBackup runs pg_basebackup for full backup
func performFullBackup(config *Config) error {
	backupDir := filepath.Join(config.Backup.FullBackupDir, time.Now().Format("20060102_150405_full"))

	// 删除指定日期之前的全量备份和增量备份
	err := cleanupOldBackups(config)
	if err != nil {
		log.Fatalf("清理旧备份失败: %v", err)
	}

	args := []string{
		"-D", backupDir, // 备份输出目录
		"-F", "tar", // 输出格式为 tar
		"-z",                       // 压缩
		"-P",                       // 显示进度
		"-U", config.Postgres.User, // 使用的 PostgreSQL 用户
		"-h", config.Postgres.Host, // 主机
		"-p", fmt.Sprintf("%d", config.Postgres.Port), // 端口
	}
	// Run pg_basebackup using the configured path
	log.Println("开始全量备份 (pg_basebackup)...")
	err = executeCommand(config.Backup.PgBasebackup, args...)
	if err != nil {
		log.Fatalf("执行全量备份失败: %v", err)
	}
	log.Println("全量备份成功完成。")

	return nil
}

// cleanupOldBackups deletes old full and incremental backups
func cleanupOldBackups(config *Config) error {
	log.Println("正在清理旧的备份数据...")

	files, err := os.ReadDir(config.Backup.FullBackupDir)
	if err != nil {
		log.Fatalf("无法读取备份目录: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			backupPath := filepath.Join(config.Backup.FullBackupDir, file.Name())
			err := os.RemoveAll(backupPath)
			if err != nil {
				log.Fatalf("无法删除备份目录 %s: %v", backupPath, err)
			}
			log.Printf("删除了旧备份: %s\n", backupPath)
		}
	}

	log.Println("旧备份数据清理完成。")
	return nil
}

func switchWAL(config *Config) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Postgres.Host, config.Postgres.Port, config.Postgres.User, config.Postgres.Password, config.Postgres.Database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// Execute the pg_switch_wal() to switch the WAL log
	_, err = db.Exec("SELECT pg_switch_wal();")
	if err != nil {
		log.Fatalf("执行 pg_switch_wal() 失败: %v", err)
	}

	log.Println("pg_switch_wal() 执行成功，WAL 日志已切换")
	return nil
}

// archiveWAL archives WAL files into a zip package
func archiveWAL(config *Config) error {
	log.Println("正在归档 WAL 文件...")

	err := switchWAL(config)
	if err != nil {
		log.Fatalf("切换 WAL 日志失败: %v", err)
	}

	srcDir := config.Backup.WalArchiveDir
	dstDir := config.Backup.WalArchiveBackupDir

	// 创建 WAL 文件压缩包名
	zipFilePath := filepath.Join(dstDir, fmt.Sprintf("wal_backup_%s.zip", time.Now().Format("20060102_150405_wal")))

	// 获取 WAL 目录中的文件列表
	files, err := os.ReadDir(srcDir)
	if err != nil {
		log.Fatalf("读取 WAL 目录失败: %v", err)
	}

	// 收集要压缩的 WAL 文件
	var walFiles []string
	for _, file := range files {
		if !file.IsDir() {
			walFiles = append(walFiles, filepath.Join(srcDir, file.Name()))
		}
	}

	// 创建 ZIP 文件并将 WAL 文件打包
	err = createZipArchive(walFiles, zipFilePath)
	if err != nil {
		log.Fatalf("创建 WAL zip 压缩包失败: %v", err)
	}

	log.Printf("WAL 文件成功打包到: %s\n", zipFilePath)
	// 当设置CleanWalArchiveDir为true时，删除WAL文件夹内的所有文件
	if config.Backup.CleanWalArchiveDir {
		for _, file := range files {
			filePath := filepath.Join(srcDir, file.Name())
			err := os.Remove(filePath)
			if err != nil {
				log.Fatalf("删除文件 %s 失败: %v", filePath, err)
			}
		}
		log.Println("WAL 目录已清空")
	}
	return nil
}

// shouldPerformFullBackup checks whether it's time for a full backup
func shouldPerformFullBackup(lastFullBackup time.Time, intervalDays int) bool {
	return time.Since(lastFullBackup).Hours() > float64(intervalDays*24)
}

// backupDatabaseWithPgDump runs pg_dump to back up the entire database in .backup format
func backupDatabaseWithPgDump(config *Config) error {
	backupDir := filepath.Join(config.Backup.FullBackupDir, time.Now().Format("20060102"))
	err := os.MkdirAll(backupDir, 0755)
	if err != nil {
		log.Fatalf("无法创建备份目录: %v", err)
	}

	backupFile := filepath.Join(backupDir, fmt.Sprintf("%s_db_full_backup.backup", config.Postgres.Database))

	args := []string{
		"-h", config.Postgres.Host,
		"-p", fmt.Sprintf("%d", config.Postgres.Port),
		"-U", config.Postgres.User,
		"-d", config.Postgres.Database, // 指定数据库
		"-F", "c", // 指定输出为 custom 格式
		"-f", backupFile, // 输出备份文件
	}
	// Run pg_dump using the configured path
	log.Printf("正在备份整个数据库为 .backup 文件: %s\n", config.Postgres.Database)
	err = executeCommand(config.Backup.PgDump, args...)
	if err != nil {
		log.Fatalf("数据库 %s 备份失败: %v", config.Postgres.Database, err)
	}
	log.Printf("数据库 %s 备份成功，备份文件路径: %s\n", config.Postgres.Database, backupFile)
	return nil
}

// backupTableWithPgDump runs pg_dump to back up specified tables in .backup format
func backupTableWithPgDump(config *Config) error {
	if len(config.Backup.Tables) == 0 {
		// 如果没有指定表，则备份整个数据库
		return backupDatabaseWithPgDump(config)
	}

	backupDir := filepath.Join(config.Backup.FullBackupDir, time.Now().Format("20060102_tables"))
	err := os.MkdirAll(backupDir, 0755)
	if err != nil {
		log.Fatalf("无法创建备份目录: %v", err)
	}

	for _, tableName := range config.Backup.Tables {
		backupFile := filepath.Join(backupDir, fmt.Sprintf("%s.backup", tableName))

		args := []string{
			"-h", config.Postgres.Host,
			"-p", fmt.Sprintf("%d", config.Postgres.Port),
			"-U", config.Postgres.User,
			"-d", config.Postgres.Database, // 指定数据库
			"-t", tableName, // 指定表名
			"-F", "c", // 指定输出为 custom 格式
			"-f", backupFile, // 输出备份文件
		}
		// Run pg_dump using the configured path
		log.Printf("正在备份表为 .backup 文件: %s\n", tableName)
		err = executeCommand(config.Backup.PgDump, args...)
		if err != nil {
			log.Fatalf("备份表 %s 失败: %v", tableName, err)
		}
		log.Printf("表 %s 备份成功，备份文件路径: %s\n", tableName, backupFile)
	}

	return nil
}

// removeOldBackups removes old backups based on the specified days
func removeOldBackups(config *Config) error {
	log.Println("正在删除旧备份文件...")

	files, err := os.ReadDir(config.Postgres.DataDir)
	if err != nil {
		log.Fatalf("无法读取备份目录: %v", err)
	}
	// 使用isFileModifyTimeOver函数判断文件是否超过指定天数，超过的话删除
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if isFileModifyTimeOver(file.Name(), config.Backup.OnlyRemainsDays) {
			filePath := filepath.Join(config.Postgres.DataDir, file.Name())
			err := os.Remove(filePath)
			if err != nil {
				log.Fatalf("删除文件 %s 失败: %v", filePath, err)
			}
			log.Printf("删除旧备份文件: %s\n", filePath)
		}
	}

	log.Println("旧备份文件删除完成。")
	return nil

}

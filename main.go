package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"os"
)

func main() {
	// 设置日志记录
	logFile, err := setupLogging()
	if err != nil {
		fmt.Printf("日志设置失败: %v\n", err)
		return
	}
	defer logFile.Close()

	// 解析命令行参数
	fullBackup := flag.Bool("f", false, "执行全量备份")
	incrementalBackup := flag.Bool("i", false, "执行增量备份")
	tableBackup := flag.Bool("t", false, "执行表备份")
	removeBackup := flag.Bool("r", false, "删除备份文件")
	flag.Parse()

	// 加载配置文件
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 设置 PGPASSWORD 环境变量
	err = os.Setenv("PGPASSWORD", config.Postgres.Password)
	if err != nil {
		log.Fatalf("设置 PGPASSWORD 环境变量失败: %v", err)
	}

	// 如果没有命令行参数，则进入常驻模式
	if !*fullBackup && !*incrementalBackup && !*tableBackup {
		c := cron.New()

		// 设置全量备份任务
		if config.Backup.FullBackupInterval > 0 {
			_, err := c.AddFunc(fmt.Sprintf("@every %dh", config.Backup.FullBackupInterval), func() {
				err := performFullBackup(config)
				if err != nil {
					log.Printf("全量备份失败: %v", err)
				}
			})
			if err != nil {
				log.Fatalf("设置全量备份任务失败: %v", err)
			}
		}

		// 设置增量备份任务
		if config.Backup.WalArchiveInterval > 0 {
			_, err := c.AddFunc(fmt.Sprintf("@every %dh", config.Backup.WalArchiveInterval), func() {
				err := archiveWAL(config)
				if err != nil {
					log.Printf("WAL 归档失败: %v", err)
				}
			})
			if err != nil {
				log.Fatalf("设置增量备份任务失败: %v", err)
			}
		}

		// 设置表备份任务
		if config.Backup.TableBackupInterval > 0 {
			_, err := c.AddFunc(fmt.Sprintf("@every %dh", config.Backup.TableBackupInterval), func() {
				err := backupTableWithPgDump(config)
				if err != nil {
					log.Printf("表备份失败: %v", err)
				}
			})
			if err != nil {
				log.Fatalf("设置表备份任务失败: %v", err)
			}
		}

		// 启动定时任务
		c.Start()

		// 保持程序运行
		select {}
	}

	// 全量备份
	if *fullBackup {
		err := performFullBackup(config)
		if err != nil {
			log.Fatalf("全量备份失败: %v", err)
		}
	}

	// 增量备份（WAL 归档）
	if *incrementalBackup {
		err := archiveWAL(config)
		if err != nil {
			log.Fatalf("WAL 归档失败: %v", err)
		}
	}

	// 单表或数据库备份
	if *tableBackup {
		err := backupTableWithPgDump(config)
		if err != nil {
			log.Fatalf("表备份失败: %v", err)
		}
	}

	// 删除指定时间前的备份文件
	if *removeBackup {
		err := removeOldBackups(config)
		if err != nil {
			log.Fatalf("删除备份文件失败: %v", err)
		}
	}
}

package main

// Config represents the structure of the config.yaml file
type Config struct {
	Postgres struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DataDir  string `yaml:"data_dir"`
		Database string `yaml:"database"`
	} `yaml:"postgres"`
	Backup struct {
		OnlyRemainsDays     int      `yaml:"only_remains_days"`
		FullBackupDir       string   `yaml:"full_backup_dir"`
		FullBackupInterval  int      `yaml:"full_backup_interval"`
		WalArchiveDir       string   `yaml:"wal_archive_dir"`
		WalArchiveBackupDir string   `yaml:"wal_archive_backup_dir"`
		CleanWalArchiveDir  bool     `yaml:"clean_wal_archive_dir"`
		WalArchiveInterval  int      `yaml:"wal_archive_interval"`
		PgBasebackup        string   `yaml:"pg_basebackup"` // pg_basebackup 的路径
		PgDump              string   `yaml:"pg_dump"`       // pg_dump 的路径
		TableBackupDir      string   `yaml:"table_backup_dir"`
		Tables              []string `yaml:"tables"`
		TableBackupInterval int      `yaml:"table_backup_interval"`
	} `yaml:"backup"`
}

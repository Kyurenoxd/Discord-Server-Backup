package main

import (
	"fmt"
	"time"
)

type BackupSchedule struct {
	Interval time.Duration
	LastRun  time.Time
}

func (bs *BackupSchedule) Start(bm *BackupManager, guildID string) {
	go func() {
		for {
			if time.Since(bs.LastRun) >= bs.Interval {
				backupName := fmt.Sprintf("auto_backup_%s", time.Now().Format("2006-01-02_15-04-05"))

				err := bm.CreateBackup(guildID, backupName)
				if err != nil {
					fmt.Printf("Scheduled backup failed: %v\n", err)
				} else {
					fmt.Printf("Created scheduled backup: %s\n", backupName)
				}
				bs.LastRun = time.Now()
			}
			time.Sleep(1 * time.Hour)
		}
	}()
}

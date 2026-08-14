package main

import (
	"log"

	"forgepack/bot"
	"forgepack/config"
	"forgepack/database"
	"forgepack/media"
	"forgepack/queue"
)

func main() {
	log.Println("Starting ForgePack...")

	cfg := config.LoadConfig()
	db := database.InitDB("bot.db", cfg.OwnerID)
	q := queue.NewProcessor(cfg.MaxQueueSize) // Max concurrent jobs
	mp := media.NewMediaProcessor(cfg.RamDiskPath)

	b := bot.InitBot(cfg, db, q, mp)
	b.Start()
}

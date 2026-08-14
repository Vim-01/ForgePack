package main

import (
	"log"

	"stickerbot/bot"
	"stickerbot/config"
	"stickerbot/database"
	"stickerbot/media"
	"stickerbot/queue"
)

func main() {
	log.Println("Starting Sticker Bot...")

	cfg := config.LoadConfig()
	db := database.InitDB("bot.db", cfg.OwnerID)
	q := queue.NewProcessor(cfg.MaxQueueSize) // Max concurrent jobs
	mp := media.NewMediaProcessor(cfg.RamDiskPath)

	b := bot.InitBot(cfg, db, q, mp)
	b.Start()
}

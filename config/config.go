package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken     string
	OwnerID      int64
	BoostByRam   bool
	RamDiskPath  string
	MaxQueueSize int
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN environment variable is required")
	}

	ownerIDStr := os.Getenv("OWNER_ID")
	var ownerID int64
	if ownerIDStr != "" {
		parsed, err := strconv.ParseInt(ownerIDStr, 10, 64)
		if err == nil {
			ownerID = parsed
		} else {
			log.Printf("Invalid OWNER_ID: %v", err)
		}
	}

	boostByRam := os.Getenv("BOOST_BY_RAM") == "true"
	ramDiskPath := os.Getenv("RAM_DISK_PATH")
	if ramDiskPath == "" {
		ramDiskPath = "/tmp/ramdisk"
	}

	maxQueueStr := os.Getenv("MAX_QUEUE_SIZE")
	maxQueueSize := 3 // default
	if maxQueueStr != "" {
		parsed, err := strconv.Atoi(maxQueueStr)
		if err == nil && parsed > 0 {
			maxQueueSize = parsed
		}
	}

	return &Config{
		BotToken:     botToken,
		OwnerID:      ownerID,
		BoostByRam:   boostByRam,
		RamDiskPath:  ramDiskPath,
		MaxQueueSize: maxQueueSize,
	}
}

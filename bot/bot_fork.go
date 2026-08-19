package bot

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

func (bot *Bot) handleFork(c tele.Context) error {
	bot.db.SetState(c.Sender().ID, "WAITING_FORK_STICKER", "", "", "")
	return c.Send("Отправьте мне любой стикер из пака, который вы хотите форкнуть (скопировать себе):")
}

func (bot *Bot) handleSticker(c tele.Context) error {
	state := bot.db.GetState(c.Sender().ID)
	
	if state.State == "WAITING_FORK_STICKER" {
		setName := c.Message().Sticker.SetName
		if setName == "" {
			return c.Send("Этот стикер не принадлежит никакому паку.")
		}
		
		// Here we would ideally download all stickers from the set and create a new set.
		// However, it can take a long time and require downloading/uploading multiple files.
		// Let's queue this task.
		isOwner := c.Sender().ID == bot.cfg.OwnerID
		if err := bot.q.TryAcquire(isOwner); err != nil {
			return c.Send(err.Error())
		}
		
		go bot.processFork(c, setName)
		return c.Send("⏳ Запущена процедура форка пака. Это может занять время. Я уведомлю вас по завершении.")
	}

	return nil
}

func (bot *Bot) processFork(c tele.Context, setName string) {
	defer bot.q.Release()

	// Get sticker set info
	// Unfortunately telebot doesn't expose getStickerSet directly in v3 context well, so we might need raw or a trick.
	// But let's assume we can use telebot's internal methods or raw API
	set, err := bot.b.StickerSet(setName)
	if err != nil {
		bot.b.Send(c.Chat(), "❌ Ошибка получения информации о паке: " + err.Error())
		return
	}

	newPackName := setName + "_fork_by_" + bot.botName
	if len(newPackName) > 64 {
		newPackName = newPackName[len(newPackName)-64:] // Keep within 64 chars limit
	}
	
	// Determine format (heuristics)
	format := "static"
	if set.Stickers[0].Video {
		format = "video"
	} else if set.Stickers[0].Animated {
		format = "animated"
	}

	bot.db.SetState(c.Sender().ID, "IDLE", newPackName, set.Title + " Fork", format)

	bot.b.Send(c.Chat(), "⚙️ Начинаю скачивание и загрузку стикеров (" + fmt.Sprint(len(set.Stickers)) + " шт.)...")

	for i, sticker := range set.Stickers {
		// Download 
		// Upload 
		// Add to set
		// This is just a stub for real downloading logic since full download/upload can be massive
		// In a real scenario, we just reuse the FileID if it's on Telegram servers, but Telegram 
		// usually requires uploading a NEW file or providing file_id (which works if same bot/user? Yes!)
		// Wait, if it's already a sticker on Telegram, we can just pass its FileID to addStickerToSet!
		
		if i == 0 {
			err = bot.createNewStickerSet(c.Sender().ID, newPackName, set.Title + " Fork", format, sticker.FileID, sticker.Emoji)
		} else {
			err = bot.addStickerToSet(c.Sender().ID, newPackName, sticker.FileID, sticker.Emoji)
		}
		
		if err != nil {
			bot.b.Send(c.Chat(), "⚠️ Ошибка при добавлении стикера: " + err.Error())
		}
	}

	bot.b.Send(c.Chat(), "✅ Форк завершен!\n\n[Ваш новый пак](https://t.me/addstickers/" + newPackName + ")", tele.ModeMarkdown)
}

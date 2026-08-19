package bot

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"forgepack/database"
	tele "gopkg.in/telebot.v3"
)

func (bot *Bot) handleMedia(c tele.Context) error {
	state := bot.db.GetState(c.Sender().ID)
	if state.State != "IDLE" {
		return c.Send("Сначала завершите текущее действие.")
	}
	if state.CurrentPackName == "" {
		return c.Send("У вас не выбран стикерпак. Используйте /newpack или /fork.")
	}

	isOwner := c.Sender().ID == bot.cfg.OwnerID
	if err := bot.q.TryAcquire(isOwner); err != nil {
		return c.Send(err.Error())
	}

	// Initial message
	msg, err := bot.b.Send(c.Chat(), "⏳ Загрузка медиа...")
	if err != nil {
		bot.q.Release()
		return err
	}

	go bot.processMedia(c, msg, state, isOwner)
	return nil
}

func (bot *Bot) processMedia(c tele.Context, msg *tele.Message, state database.UserState, isOwner bool) {
	defer bot.q.Release()

	var fileID string
	var isImage bool
	
	if c.Message().Photo != nil {
		fileID = c.Message().Photo.FileID
		isImage = true
	} else if c.Message().Video != nil {
		fileID = c.Message().Video.FileID
		isImage = false
	} else if c.Message().Document != nil {
		fileID = c.Message().Document.FileID
		isImage = strings.HasPrefix(c.Message().Document.MIME, "image/")
	} else {
		c.Send("Неподдерживаемый формат.")
		return
	}

	caption := c.Message().Caption
	removeBg := strings.Contains(caption, "-B") || strings.Contains(caption, "--Background")

	// 1. Download File
	file, err := bot.b.FileByID(fileID)
	if err != nil {
		bot.b.Edit(msg, "❌ Ошибка получения файла")
		return
	}
	
	tmpDir, _ := os.MkdirTemp("", "sticker_*")
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "input")
	if err := bot.downloadFile(file, inPath); err != nil {
		bot.b.Edit(msg, "❌ Ошибка скачивания")
		return
	}

	bot.b.Edit(msg, "⚙️ Обработка медиа...")

	outPath := filepath.Join(tmpDir, "output")
	
	// Media conversions
	if state.CurrentPackType == "static" {
		if !isImage {
			bot.b.Send(c.Chat(), "⚠️ Внимание: Вы загрузили видео в статический пак. Будет извлечен только первый кадр.")
		}
		
		outPath += ".webp"
		if removeBg {
			if isImage {
				bot.media.RemoveBackgroundStatic(inPath, outPath)
				bot.media.FormatToStatic(outPath, outPath) // Resize
			} else {
				// extract frame first, then remove bg
				tmpFrame := filepath.Join(tmpDir, "frame.webp")
				bot.media.FormatToStatic(inPath, tmpFrame)
				bot.media.RemoveBackgroundStatic(tmpFrame, outPath)
			}
		} else {
			if err := bot.media.FormatToStatic(inPath, outPath); err != nil {
				bot.handleErr(msg, "Ошибка конвертации", err)
				return
			}
		}
	} else if state.CurrentPackType == "video" {
		outPath += ".webm"
		if removeBg {
			bot.b.Edit(msg, "🧠 Нейросеть удаляет фон (это может занять время)...")
			if err := bot.media.RemoveBackgroundVideo(inPath, outPath); err != nil {
				bot.handleErr(msg, "Ошибка удаления фона видео", err)
				return
			}
		} else {
			if err := bot.media.FormatToVideo(inPath, outPath, isImage); err != nil {
				bot.handleErr(msg, "Ошибка конвертации видео", err)
				return
			}
		}
	}

	bot.b.Edit(msg, "🚀 Отправка в Telegram...")

	formatStr := "static"
	if state.CurrentPackType == "video" {
		formatStr = "video"
	}
	
	uploadedFileID, err := bot.uploadStickerFile(c.Sender().ID, outPath, formatStr)
	if err != nil {
		bot.handleErr(msg, "Ошибка загрузки файла в Telegram", err)
		return
	}

	bot.b.Edit(msg, "📦 Добавление в пак...")
	
	// Assuming default emoji is 🙂
	emoji := "🙂" 

	// Try to add, if fails with "STICKERSET_INVALID", maybe create it?
	err = bot.addStickerToSet(c.Sender().ID, state.CurrentPackName, uploadedFileID, emoji)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "not found") {
			// Try to create set
			err = bot.createNewStickerSet(c.Sender().ID, state.CurrentPackName, state.CurrentPackTitle, formatStr, uploadedFileID, emoji)
			if err != nil {
				bot.handleErr(msg, "Ошибка создания пака", err)
				return
			}
		} else {
			bot.handleErr(msg, "Ошибка добавления стикера", err)
			return
		}
	}

	bot.b.Edit(msg, fmt.Sprintf("✅ Стикер успешно добавлен в пак!\n\n[Ваш пак](https://t.me/addstickers/%s)", state.CurrentPackName), tele.ModeMarkdown)
	return
}

func (bot *Bot) handleErr(msg *tele.Message, text string, err error) error {
	bot.b.Edit(msg, fmt.Sprintf("❌ %s: %v", text, err))
	return err
}

func (bot *Bot) downloadFile(file tele.File, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	rc, err := bot.b.File(&file)
	if err != nil {
		return err
	}
	defer rc.Close()

	_, err = io.Copy(out, rc)
	return err
}

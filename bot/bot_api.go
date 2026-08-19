package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func (bot *Bot) uploadStickerFile(userID int64, filePath string, stickerFormat string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("user_id", fmt.Sprintf("%d", userID))
	_ = writer.WriteField("sticker_format", stickerFormat)
	
	part, _ := writer.CreateFormFile("sticker", filePath)
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://api.telegram.org/bot%s/uploadStickerFile", bot.cfg.BotToken), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		Ok     bool `json:"ok"`
		Result struct {
			FileID string `json:"file_id"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Ok {
		return "", fmt.Errorf("failed to upload sticker file")
	}
	return res.Result.FileID, nil
}

func (bot *Bot) createNewStickerSet(userID int64, name, title, format string, fileID string, emoji string) error {
	type InputSticker struct {
		Sticker   string   `json:"sticker"`
		EmojiList []string `json:"emoji_list"`
	}
	
	stickers := []InputSticker{{
		Sticker:   fileID,
		EmojiList: []string{emoji},
	}}
	
	stickersJSON, _ := json.Marshal(stickers)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("user_id", fmt.Sprintf("%d", userID))
	_ = writer.WriteField("name", name)
	_ = writer.WriteField("title", title)
	_ = writer.WriteField("sticker_format", format)
	_ = writer.WriteField("stickers", string(stickersJSON))
	writer.Close()

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://api.telegram.org/bot%s/createNewStickerSet", bot.cfg.BotToken), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var res struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Ok {
		return fmt.Errorf("failed to create set: %s", res.Description)
	}
	return nil
}

func (bot *Bot) addStickerToSet(userID int64, name string, fileID string, emoji string) error {
	type InputSticker struct {
		Sticker   string   `json:"sticker"`
		EmojiList []string `json:"emoji_list"`
	}
	
	sticker := InputSticker{
		Sticker:   fileID,
		EmojiList: []string{emoji},
	}
	
	stickerJSON, _ := json.Marshal(sticker)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("user_id", fmt.Sprintf("%d", userID))
	_ = writer.WriteField("name", name)
	_ = writer.WriteField("sticker", string(stickerJSON))
	writer.Close()

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://api.telegram.org/bot%s/addStickerToSet", bot.cfg.BotToken), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var res struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	if !res.Ok {
		return fmt.Errorf("failed to add sticker: %s", res.Description)
	}
	return nil
}

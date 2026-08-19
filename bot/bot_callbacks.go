package bot

import (
	tele "gopkg.in/telebot.v3"
)

func (bot *Bot) initCallbacks() {
	bot.b.Handle(tele.OnCallback, bot.handleCallback)
}

func (bot *Bot) handleCallback(c tele.Context) error {
	data := c.Callback().Data
	state := bot.db.GetState(c.Sender().ID)

	if state.State == "WAITING_PACK_TYPE" {
		packType := "static"
		if data == "type_video" {
			packType = "video"
		}
		
		bot.db.SetState(c.Sender().ID, "IDLE", state.CurrentPackName, state.CurrentPackTitle, packType)
		c.Edit("Тип выбран: " + packType + "\n\nТеперь вы можете отправлять мне медиафайлы для создания первого стикера в этом паке (пак будет создан при отправке первого стикера).")
		return c.Respond()
	}
	
	return c.Respond()
}

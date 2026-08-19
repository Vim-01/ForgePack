package bot

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"forgepack/config"
	"forgepack/database"
	"forgepack/media"
	"forgepack/queue"

	tele "gopkg.in/telebot.v3"
)

type Bot struct {
	b       *tele.Bot
	db      *database.DB
	cfg     *config.Config
	q       *queue.Processor
	media   *media.MediaProcessor
	botName string
}

func InitBot(cfg *config.Config, db *database.DB, q *queue.Processor, mp *media.MediaProcessor) *Bot {
	pref := tele.Settings{
		Token:  cfg.BotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	bot := &Bot{
		b:       b,
		db:      db,
		cfg:     cfg,
		q:       q,
		media:   mp,
		botName: b.Me.Username,
	}

	bot.setupHandlers()
	return bot
}

func (bot *Bot) Start() {
	log.Println("Bot started as", bot.botName)
	bot.b.Start()
}

func (bot *Bot) checkAccess(c tele.Context) bool {
	if c.Sender().ID == bot.cfg.OwnerID {
		return true
	}
	return bot.db.IsAllowed(c.Sender().ID)
}

func (bot *Bot) setupHandlers() {
	bot.b.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if !bot.checkAccess(c) {
				return c.Send("У вас нет доступа к этому боту. Обратитесь к владельцу.")
			}
			return next(c)
		}
	})

	bot.b.Handle("/start", bot.handleStart)
	bot.b.Handle("/add_user", bot.handleAddUser)
	bot.b.Handle("/newpack", bot.handleNewPack)
	bot.b.Handle("/fork", bot.handleFork)

	bot.b.Handle(tele.OnText, bot.handleText)
	bot.b.Handle(tele.OnPhoto, bot.handleMedia)
	bot.b.Handle(tele.OnVideo, bot.handleMedia)
	bot.b.Handle(tele.OnDocument, bot.handleMedia)
	bot.b.Handle(tele.OnSticker, bot.handleSticker)
}

func (bot *Bot) handleStart(c tele.Context) error {
	msg := "Привет! Я бот для создания стикерпаков.\n\n" +
		"Доступные команды:\n" +
		"`/newpack` - создать новый пак\n" +
		"`/fork` - форкнуть чужой пак\n\n" +
		"Отправьте мне фото или видео для добавления в текущий пак.\n" +
		"Добавьте подпись `-B`, чтобы удалить фон!"
	return c.Send(msg, tele.ModeMarkdownV2)
}

func (bot *Bot) handleAddUser(c tele.Context) error {
	if c.Sender().ID != bot.cfg.OwnerID {
		return c.Send("Только владелец может добавлять пользователей.")
	}
	args := c.Args()
	if len(args) != 1 {
		return c.Send("Использование: /add_user <ID>")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return c.Send("Неверный ID.")
	}
	bot.db.AllowUser(id)
	return c.Send(fmt.Sprintf("Пользователь %d добавлен.", id))
}

func (bot *Bot) handleNewPack(c tele.Context) error {
	bot.db.SetState(c.Sender().ID, "WAITING_PACK_TITLE", "", "", "")
	return c.Send("Введите название (Title) для нового стикерпака:")
}

func (bot *Bot) handleText(c tele.Context) error {
	state := bot.db.GetState(c.Sender().ID)
	text := c.Text()

	switch state.State {
	case "WAITING_PACK_TITLE":
		bot.db.SetState(c.Sender().ID, "WAITING_PACK_NAME", "", text, "")
		return c.Send("Теперь введите короткое имя (Name) для ссылки (только англ буквы, цифры, без пробелов). Суффикс бота добавится автоматически:")
	case "WAITING_PACK_NAME":
		name := text + "_by_" + bot.botName
		bot.db.SetState(c.Sender().ID, "WAITING_PACK_TYPE", name, state.CurrentPackTitle, "")
		
		btnStatic := tele.InlineButton{Text: "Статический", Data: "type_static"}
		btnVideo := tele.InlineButton{Text: "Видео", Data: "type_video"}
		markup := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnStatic, btnVideo}}}
		return c.Send("Выберите тип стикерпака:", markup)
	}
	return nil
}

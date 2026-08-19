<h1 align="center">
  <br>
  StickerBot 🤖🎨
  <br>
</h1>

<h4 align="center"> Telegram-бот для создания, форка и редактирования стикерпаков (Photo & Video)</h4>

<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Python-3776AB?style=for-the-badge&logo=python&logoColor=white" alt="Python">
  <img src="https://img.shields.io/badge/Docker-2CA5E0?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/FFmpeg-007808?style=for-the-badge&logo=ffmpeg&logoColor=white" alt="FFmpeg">
  <img src="https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white" alt="Telegram API">
</p>

<p align="center">
  <a href="#о-проекте">О проекте</a> •
  <a href="#возможности">Возможности</a> •
  <a href="#настройка-и-запуск">Настройка и Запуск</a> •
  <a href="#использование">Использование</a>
</p>

---

## 💡 О проекте

**StickerBot** — это высокопроизводительный self-hosted Telegram-бот, на **Go** с использованием интеграции легковесных ai. 
Бот ориентирован на миниальное потребление рессурсов. Имеет очереди в работе с пользователями для снижения рисков.

## ✨ Возможности

- 🚀 **Производительность:** Ядро на Golang обеспечивает сверхбыструю обработку сообщений и конкурентность.
- 🎨 **Удаление фона:** Интегрированная нейросеть `U-2-Net` (через `rembg`) позволяет стирать фон как на фото, так и на видео (покадрово). Просто добавьте флаг `-B` к вашему медиа!
- 🔄 **Автоконвертация:** 
  - Загрузили видео в статический пак? Бот аккуратно извлечет первый кадр.
  - Загрузили фото в видеопак? Бот превратит его в 3-секундное видео.
  - Любые фото и видео автоматически приводятся к жестким стандартам Telegram.
- 🔀 **Форк Стикерпаков:** Команда `/fork` полностью скопирует его стикеры в ваш личный пак для дальнейшего редактирования.
- 🛡 **Защита сервера (Queue Limits):** Встроенная очередь настраивается через `.env` . Все лишние запросы вежливо отклоняются, пока сервер не освободится. Владелец бота (`OWNER_ID`) имеет иммунитет к очереди.
- ⚡ **RAM Boost (tmpfs):** Поддержка монтирования временной папки в ОЗУ для обработки кадров видео, что сильно экономит ресурс SSD/HDD на слабых VDS.

## 🛠 Настройка и Запуск

Проект полностью упакован в `Docker` для удобства.

### Требования
- ~ 1 gb ram VDS. 
-  `Docker` и `Docker Compose`

### Установка

1. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/Vim-01/stickerbot.git
   cd stickerbot
   ```

2. Скопируйте файл конфигурации:
   ```bash
   cp .env.example .env
   ```

3. Отредактируйте `.env`, вставив ваши данные:
   ```ini
   BOT_TOKEN=123456789:ABCDefghIJKlmnOPQRSTuvwxyz # Токен от @BotFather
   OWNER_ID=123456789 # Ваш Telegram ID (получить через @ShowJsonBot или аналогичный)
   MAX_QUEUE_SIZE=3 # Максимум тяжелых задач одновременно
   BOOST_BY_RAM=true # Использовать ОЗУ для кэширования
   RAM_DISK_PATH=/tmp/ramdisk # Путь монтирования tmpfs внутри контейнера
   ```

4. *(Опционально)* Если вы включили `BOOST_BY_RAM=true`, раскомментируйте блок `tmpfs` в файле `docker-compose.yml` .

5. Запустите бота:
   ```bash
   docker-compose up -d --build
   ```

## 🎮 Использование

Вы можете выдавать доступ к боту своим друзьям командой:
`/add_user <Telegram ID>`

### Основные команды бота:
* `/newpack` — Создать новый пустой стикерпак.
* `/fork` — Скопировать любой существующий чужой пак к себе.
* `/start` — Помощь и информация.

### Модификаторы сообщений (Подписи)
Когда бот находится в режиме ожидания нового стикера, отправьте ему фото или видео.
* Прикрепите подпись **`-B`** или **`--Background`** к медиафайлу, чтобы бот автоматически удалил с него задний фон! 

---
*With <3*

# yt2anki

CLI-утилита для создания Anki-колод из YouTube видео. Извлекает лексику уровня A2-B2 из транскрипции видео с помощью LLM.

## Требования

### Внешние зависимости

1. **yt-dlp** — скачивание аудио с YouTube

```bash
# macOS
brew install yt-dlp

# Linux
pip install yt-dlp

# Windows
winget install yt-dlp
```

2. **mlx_whisper** — транскрипция аудио (Whisper для Apple Silicon)

```bash
pip install mlx-whisper
```

> Требуется Mac с Apple Silicon (M1/M2/M3/M4)

3. **ffmpeg** — требуется для yt-dlp и whisper

```bash
# macOS
brew install ffmpeg

# Linux (Ubuntu/Debian)
sudo apt install ffmpeg

# Windows
winget install ffmpeg
```

### LLM API

Нужен доступ к OpenAI-compatible API:

- **OpenAI API** — установите `OPENAI_API_KEY`
- **Ollama** (локально) — запустите `ollama serve` и используйте `--api-url http://localhost:11434/v1`
- **Любой OpenAI-compatible сервер** — укажите URL через `--api-url`

## Установка

```bash
# Клонирование
git clone https://github.com/weazyexe/yt2anki.git
cd yt2anki

# Сборка
go build -o yt2anki .

# Опционально: установка в PATH
sudo mv yt2anki /usr/local/bin/
```

## Использование

```bash
yt2anki [flags] <youtube-url>

Flags:
  -n, --count int       Количество слов (default: 20)
  -o, --output string   Выходной файл (default: deck.apkg)
  -l, --level string    Уровень: A2, B1, B2 (default: B1)
      --api-url string  OpenAI-compatible API URL (default: http://localhost:11434/v1)
      --api-key string  API ключ (или env: OPENAI_API_KEY)
      --model string    Модель LLM (default: gpt-4o-mini)
```

### Примеры

```bash
# Базовый запуск с OpenAI
export OPENAI_API_KEY=sk-xxx
yt2anki "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# 30 слов уровня B2
yt2anki -n 30 -l B2 -o english_vocab.apkg "https://youtu.be/xxx"

# С локальным Ollama
yt2anki --api-url http://localhost:11434/v1 --model llama3 --api-key dummy "URL"
```

## Структура карточек

Каждое слово создаёт 2 карточки:

**Forward (EN → RU):**
```
Front: Word
Back:  Definition (RU)
       /IPA/
       Example (EN)
       Example (RU)
```

**Reverse (RU → EN):**
```
Front: Definition (RU)
Back:  Word
       /IPA/
       Example (EN)
       Example (RU)
```

## Импорт в Anki

1. Откройте Anki
2. File → Import
3. Выберите сгенерированный `.apkg` файл
4. Готово

## Troubleshooting

**"yt-dlp not found"** — установите yt-dlp и убедитесь, что он доступен в PATH

**"mlx_whisper not found"** — установите mlx-whisper: `pip install mlx-whisper`

**"API key required"** — установите переменную окружения `OPENAI_API_KEY` или используйте флаг `--api-key`

**Долгая транскрипция** — whisper использует модель `medium`, для длинных видео это может занять время. Убедитесь, что у вас достаточно RAM (рекомендуется 8GB+)

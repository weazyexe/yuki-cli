# yt2anki — Техническое задание

## Цель

CLI-утилита на Go: YouTube видео → транскрипция → извлечение B1-B2 лексики через LLM → Anki колода (.apkg)

## Пайплайн

```
YouTube URL → yt-dlp (mp3) → whisper (transcript) → LLM (vocabulary) → APKG
```

## Зависимости (внешние)

- `yt-dlp` — скачивание аудио
- `whisper` — транскрипция (модель: medium)
- OpenAI-compatible API — анализ лексики

## Структура карточки

### Поля

| Поле         | Описание                    |
| ------------ | --------------------------- |
| `Word`       | Слово или фраза (EN)        |
| `Definition` | Определение на русском      |
| `IPA`        | Транскрипция (фонетическая) |
| `ExampleEN`  | Пример предложения (EN)     |
| `ExampleRU`  | Перевод примера (RU)        |

### Типы карточек

**Forward (EN → RU):**

```
Front: {{Word}}
Back:  {{Definition}}
       /{{IPA}}/
       {{ExampleEN}}
       {{ExampleRU}}
```

**Reverse (RU → EN):**

```
Front: {{Definition}}
Back:  {{Word}}
       /{{IPA}}/
       {{ExampleEN}}
       {{ExampleRU}}
```

## CLI интерфейс

```bash
yt2anki [flags] <youtube-url>

Flags:
  -n, --count int       Количество слов (default: 20)
  -o, --output string   Выходной файл (default: deck.apkg)
  -l, --level string    Уровень: A2, B1, B2 (default: B1)
  --api-url string      OpenAI-compatible API URL (default: http://localhost:11434/v1)
  --api-key string      API ключ (или env: OPENAI_API_KEY)
  --model string        Модель LLM (default: gpt-4o-mini)
```

## LLM промпт (упрощённо)

```
Из транскрипта выбери {{count}} слов/фраз уровня {{level}}.

Для каждого верни JSON:
{
  "word": "string",
  "definition": "string (на русском)",
  "ipa": "string",
  "example_en": "string",
  "example_ru": "string"
}

Транскрипт:
{{transcript}}
```

## Структура проекта

```
yt2anki/
├── main.go              # CLI entry point
├── internal/
│   ├── downloader.go    # yt-dlp wrapper
│   ├── transcriber.go   # whisper wrapper
│   ├── llm.go           # OpenAI client
│   └── anki.go          # APKG generator
├── go.mod
└── go.sum
```

## Библиотеки Go

- `github.com/spf13/cobra` — CLI
- `github.com/sashabaranov/go-openai` — OpenAI client
- ZIP archive/генерация APKG — стандартная библиотека + SQLite

## Формат APKG

APKG = ZIP архив содержащий:

- `collection.anki2` — SQLite база с таблицами: `col`, `notes`, `cards`
- `media` — JSON маппинг медиафайлов (пустой для нашего случая)

## Порядок реализации

1. CLI skeleton (cobra)
2. Downloader: `yt-dlp -x --audio-format mp3`
3. Transcriber: `whisper --model medium --output_format txt`
4. LLM client: POST /chat/completions
5. APKG generator: SQLite + ZIP
6. Интеграция + cleanup temp files

## Пример использования

```bash
# Базовый запуск
export OPENAI_API_KEY=sk-xxx
yt2anki "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# С параметрами
yt2anki -n 30 -l B2 -o english_vocab.apkg "https://youtu.be/xxx"

# Локальный LLM (ollama)
yt2anki --api-url http://localhost:11434/v1 --model llama3 "URL"
```

## Обработка ошибок

- Проверка наличия yt-dlp, whisper в PATH
- Валидация YouTube URL
- Таймаут для LLM запросов (60s)
- Cleanup временных файлов при любом исходе

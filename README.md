# yuki

CLI-инструмент для создания Anki-карточек из YouTube видео или файлов субтитров.

## Возможности

- Извлечение словаря из YouTube видео (скачивание + транскрибирование)
- Поддержка файлов субтитров (SRT, VTT) и текстовых файлов (TXT)
- Извлечение слов по уровням CEFR (A2, B1, B2)
- Интерактивный выбор слов для колоды
- Кеширование аудио и транскриптов
- Генерация готовых .apkg файлов для Anki

## Установка

### Зависимости

Для работы с YouTube видео требуются:

- [yt-dlp](https://github.com/yt-dlp/yt-dlp) — скачивание аудио
- [mlx_whisper](https://github.com/ml-explore/mlx-examples/tree/main/whisper) — транскрибирование (Apple Silicon)

```bash
# macOS
brew install yt-dlp
pip install mlx-whisper
```

### Сборка

```bash
git clone https://github.com/weazyexe/yuki-cli.git
cd yuki-cli
go build -o yuki .
```

## Использование

### Из YouTube видео

```bash
# Базовое использование
yuki https://www.youtube.com/watch?v=VIDEO_ID

# С параметрами
yuki -n 30 -l B2 -o vocabulary.apkg https://youtu.be/VIDEO_ID
```

### Из файла субтитров

```bash
# SRT файл
yuki subtitles.srt

# VTT файл
yuki captions.vtt

# Текстовый файл
yuki transcript.txt
```

При использовании файлов зависимости yt-dlp и mlx_whisper не требуются.

## Флаги

| Флаг            | Короткий | По умолчанию       | Описание                            |
| --------------- | -------- | ------------------ | ----------------------------------- |
| `--count`       | `-n`     | 20                 | Количество слов для извлечения      |
| `--output`      | `-o`     | deck.apkg          | Путь к выходному файлу              |
| `--level`       | `-l`     | B1                 | Уровень языка: A2, B1, B2           |
| `--api-url`     |          | localhost:11434/v1 | URL OpenAI-совместимого API         |
| `--api-key`     |          |                    | API ключ (или env: OPENAI_API_KEY)  |
| `--model`       |          | gpt-4o-mini        | Название LLM модели                 |
| `--no-review`   |          | false              | Пропустить интерактивный выбор слов |
| `--no-cache`    |          | false              | Отключить кеширование               |
| `--refresh`     |          | false              | Игнорировать кеш, скачать заново    |
| `--clear-cache` |          |                    | Очистить кеш и выйти                |

## Примеры

```bash
# Извлечь 30 слов уровня B2, пропустить выбор
yuki -n 30 -l B2 --no-review https://www.youtube.com/watch?v=abc123

# Использовать OpenAI API
yuki --api-url https://api.openai.com/v1 \
        --api-key $OPENAI_API_KEY \
        --model gpt-4o \
        video.srt

# Использовать локальный Ollama
yuki --api-url http://localhost:11434/v1 \
        --api-key ollama \
        --model llama3.2 \
        transcript.txt
```

## Кеширование

Кеш хранится в `~/.cache/yuki/` (или `$XDG_CACHE_HOME/yuki/`):

- `audio/` — скачанные аудиофайлы
- `transcripts/` — транскрипты

```bash
# Очистить кеш
yuki --clear-cache

# Игнорировать кеш для одного запуска
yuki --refresh https://youtube.com/...

# Отключить кеширование
yuki --no-cache https://youtube.com/...
```

## Поддерживаемые форматы

| Формат | Расширение | Описание             |
| ------ | ---------- | -------------------- |
| SubRip | .srt       | Стандартные субтитры |
| WebVTT | .vtt       | Веб-субтитры         |
| Text   | .txt       | Обычный текст        |

## Тестирование

```bash
go test ./...
```

## Лицензия

MIT

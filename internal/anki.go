package internal

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// Model and deck IDs (generated once, consistent for the app)
	modelID = 1704067200000
	deckID  = 1704067200001
)

// Card templates
const (
	frontTemplate = `<div class="word">{{Word}}</div>`

	backTemplate = `<div class="word">{{FrontSide}}</div>
<hr id="answer">
<div class="definition">{{Definition}}</div>
<div class="ipa">/{{IPA}}/</div>
<div class="example">
  <div class="en">{{ExampleEN}}</div>
  <div class="ru">{{ExampleRU}}</div>
</div>`

	reverseFrontTemplate = `<div class="definition">{{Definition}}</div>`

	reverseBackTemplate = `<div class="definition">{{FrontSide}}</div>
<hr id="answer">
<div class="word">{{Word}}</div>
<div class="ipa">/{{IPA}}/</div>
<div class="example">
  <div class="en">{{ExampleEN}}</div>
  <div class="ru">{{ExampleRU}}</div>
</div>`

	css = `.card {
  font-family: arial;
  font-size: 20px;
  text-align: center;
  color: black;
  background-color: white;
}
.word {
  font-size: 28px;
  font-weight: bold;
  color: #2196F3;
}
.definition {
  font-size: 22px;
  margin: 10px 0;
}
.ipa {
  font-size: 18px;
  color: #666;
  font-style: italic;
}
.example {
  margin-top: 15px;
  text-align: left;
  padding: 10px;
  background: #f5f5f5;
  border-radius: 5px;
}
.example .en {
  font-weight: bold;
}
.example .ru {
  color: #666;
  margin-top: 5px;
}`
)

// GenerateAPKG creates an Anki package file from vocabulary items
func GenerateAPKG(items []VocabularyItem, outputPath, deckName string) error {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "anki-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create SQLite database
	dbPath := filepath.Join(tempDir, "collection.anki2")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer db.Close()

	if err := initializeDatabase(db, deckName); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := insertNotes(db, items); err != nil {
		return fmt.Errorf("failed to insert notes: %w", err)
	}

	db.Close()

	// Create APKG (ZIP archive)
	if err := createAPKG(tempDir, outputPath); err != nil {
		return fmt.Errorf("failed to create APKG: %w", err)
	}

	return nil
}

func initializeDatabase(db *sql.DB, deckName string) error {
	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS col (
		id INTEGER PRIMARY KEY,
		crt INTEGER NOT NULL,
		mod INTEGER NOT NULL,
		scm INTEGER NOT NULL,
		ver INTEGER NOT NULL,
		dty INTEGER NOT NULL,
		usn INTEGER NOT NULL,
		ls INTEGER NOT NULL,
		conf TEXT NOT NULL,
		models TEXT NOT NULL,
		decks TEXT NOT NULL,
		dconf TEXT NOT NULL,
		tags TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY,
		guid TEXT NOT NULL,
		mid INTEGER NOT NULL,
		mod INTEGER NOT NULL,
		usn INTEGER NOT NULL,
		tags TEXT NOT NULL,
		flds TEXT NOT NULL,
		sfld TEXT NOT NULL,
		csum INTEGER NOT NULL,
		flags INTEGER NOT NULL,
		data TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS cards (
		id INTEGER PRIMARY KEY,
		nid INTEGER NOT NULL,
		did INTEGER NOT NULL,
		ord INTEGER NOT NULL,
		mod INTEGER NOT NULL,
		usn INTEGER NOT NULL,
		type INTEGER NOT NULL,
		queue INTEGER NOT NULL,
		due INTEGER NOT NULL,
		ivl INTEGER NOT NULL,
		factor INTEGER NOT NULL,
		reps INTEGER NOT NULL,
		lapses INTEGER NOT NULL,
		left INTEGER NOT NULL,
		odue INTEGER NOT NULL,
		odid INTEGER NOT NULL,
		flags INTEGER NOT NULL,
		data TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS revlog (
		id INTEGER PRIMARY KEY,
		cid INTEGER NOT NULL,
		usn INTEGER NOT NULL,
		ease INTEGER NOT NULL,
		ivl INTEGER NOT NULL,
		lastIvl INTEGER NOT NULL,
		factor INTEGER NOT NULL,
		time INTEGER NOT NULL,
		type INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS graves (
		usn INTEGER NOT NULL,
		oid INTEGER NOT NULL,
		type INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS ix_notes_usn ON notes (usn);
	CREATE INDEX IF NOT EXISTS ix_cards_usn ON cards (usn);
	CREATE INDEX IF NOT EXISTS ix_revlog_usn ON revlog (usn);
	CREATE INDEX IF NOT EXISTS ix_cards_nid ON cards (nid);
	CREATE INDEX IF NOT EXISTS ix_cards_sched ON cards (did, queue, due);
	CREATE INDEX IF NOT EXISTS ix_revlog_cid ON revlog (cid);
	CREATE INDEX IF NOT EXISTS ix_notes_csum ON notes (csum);
	`

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Insert collection metadata
	now := time.Now().Unix()
	models := createModels()
	decks := createDecks(deckName)
	conf := createConf()
	dconf := createDconf()

	modelsJSON, _ := json.Marshal(models)
	decksJSON, _ := json.Marshal(decks)
	confJSON, _ := json.Marshal(conf)
	dconfJSON, _ := json.Marshal(dconf)

	_, err := db.Exec(`
		INSERT INTO col (id, crt, mod, scm, ver, dty, usn, ls, conf, models, decks, dconf, tags)
		VALUES (1, ?, ?, ?, 11, 0, -1, 0, ?, ?, ?, ?, '{}')
	`, now, now*1000, now*1000, string(confJSON), string(modelsJSON), string(decksJSON), string(dconfJSON))

	return err
}

func createModels() map[string]interface{} {
	mid := fmt.Sprintf("%d", modelID)

	return map[string]interface{}{
		mid: map[string]interface{}{
			"id":    modelID,
			"name":  "yt2anki Vocabulary",
			"type":  0,
			"mod":   time.Now().Unix(),
			"usn":   -1,
			"sortf": 0,
			"did":   deckID,
			"tmpls": []map[string]interface{}{
				{
					"name":  "Forward (EN → RU)",
					"qfmt":  frontTemplate,
					"afmt":  backTemplate,
					"bqfmt": "",
					"bafmt": "",
					"ord":   0,
					"did":   nil,
				},
				{
					"name":  "Reverse (RU → EN)",
					"qfmt":  reverseFrontTemplate,
					"afmt":  reverseBackTemplate,
					"bqfmt": "",
					"bafmt": "",
					"ord":   1,
					"did":   nil,
				},
			},
			"flds": []map[string]interface{}{
				{"name": "Word", "ord": 0, "sticky": false, "rtl": false, "font": "Arial", "size": 20, "media": []string{}},
				{"name": "Definition", "ord": 1, "sticky": false, "rtl": false, "font": "Arial", "size": 20, "media": []string{}},
				{"name": "IPA", "ord": 2, "sticky": false, "rtl": false, "font": "Arial", "size": 20, "media": []string{}},
				{"name": "ExampleEN", "ord": 3, "sticky": false, "rtl": false, "font": "Arial", "size": 20, "media": []string{}},
				{"name": "ExampleRU", "ord": 4, "sticky": false, "rtl": false, "font": "Arial", "size": 20, "media": []string{}},
			},
			"css":  css,
			"latexPre": `\documentclass[12pt]{article}
\special{papersize=3in,5in}
\usepackage[utf8]{inputenc}
\usepackage{amssymb,amsmath}
\pagestyle{empty}
\setlength{\parindent}{0in}
\begin{document}`,
			"latexPost": `\end{document}`,
			"latexsvg":  false,
			"req":       [][]interface{}{{0, "any", []int{0}}, {1, "any", []int{1}}},
		},
	}
}

func createDecks(deckName string) map[string]interface{} {
	did := fmt.Sprintf("%d", deckID)

	return map[string]interface{}{
		"1": map[string]interface{}{
			"id":             1,
			"name":           "Default",
			"mod":            time.Now().Unix(),
			"usn":            -1,
			"lrnToday":       []int{0, 0},
			"revToday":       []int{0, 0},
			"newToday":       []int{0, 0},
			"timeToday":      []int{0, 0},
			"collapsed":      false,
			"browserCollapsed": false,
			"desc":           "",
			"dyn":            0,
			"conf":           1,
		},
		did: map[string]interface{}{
			"id":             deckID,
			"name":           deckName,
			"mod":            time.Now().Unix(),
			"usn":            -1,
			"lrnToday":       []int{0, 0},
			"revToday":       []int{0, 0},
			"newToday":       []int{0, 0},
			"timeToday":      []int{0, 0},
			"collapsed":      false,
			"browserCollapsed": false,
			"desc":           "Vocabulary deck created by yt2anki",
			"dyn":            0,
			"conf":           1,
		},
	}
}

func createConf() map[string]interface{} {
	return map[string]interface{}{
		"activeDecks":   []int{1},
		"curDeck":       1,
		"newSpread":     0,
		"collapseTime":  1200,
		"timeLim":       0,
		"estTimes":      true,
		"dueCounts":     true,
		"curModel":      nil,
		"nextPos":       1,
		"sortType":      "noteFld",
		"sortBackwards": false,
		"addToCur":      true,
	}
}

func createDconf() map[string]interface{} {
	return map[string]interface{}{
		"1": map[string]interface{}{
			"id":       1,
			"name":     "Default",
			"mod":      0,
			"usn":      0,
			"maxTaken": 60,
			"autoplay": true,
			"timer":    0,
			"replayq":  true,
			"new": map[string]interface{}{
				"bury":      true,
				"delays":    []float64{1, 10},
				"initialFactor": 2500,
				"ints":      []int{1, 4, 7},
				"order":     1,
				"perDay":    20,
			},
			"rev": map[string]interface{}{
				"bury":     true,
				"ease4":    1.3,
				"fuzz":     0.05,
				"ivlFct":   1,
				"maxIvl":   36500,
				"perDay":   200,
				"hardFactor": 1.2,
			},
			"lapse": map[string]interface{}{
				"delays":    []float64{10},
				"leechAction": 0,
				"leechFails": 8,
				"minInt":    1,
				"mult":      0,
			},
			"dyn":   false,
		},
	}
}

func insertNotes(db *sql.DB, items []VocabularyItem) error {
	now := time.Now().Unix()
	cardID := now * 1000

	for i, item := range items {
		noteID := now*1000 + int64(i)
		guid := fmt.Sprintf("yt2anki%d", noteID)

		// Fields separated by \x1f (unit separator)
		fields := fmt.Sprintf("%s\x1f%s\x1f%s\x1f%s\x1f%s",
			html.EscapeString(item.Word),
			html.EscapeString(item.Definition),
			html.EscapeString(item.IPA),
			html.EscapeString(item.ExampleEN),
			html.EscapeString(item.ExampleRU),
		)

		// Simple checksum based on first field
		csum := fieldChecksum(item.Word)

		// Insert note
		_, err := db.Exec(`
			INSERT INTO notes (id, guid, mid, mod, usn, tags, flds, sfld, csum, flags, data)
			VALUES (?, ?, ?, ?, -1, '', ?, ?, ?, 0, '')
		`, noteID, guid, modelID, now, fields, item.Word, csum)
		if err != nil {
			return fmt.Errorf("failed to insert note: %w", err)
		}

		// Insert cards (2 cards per note: forward and reverse)
		for ord := 0; ord < 2; ord++ {
			cardID++
			_, err := db.Exec(`
				INSERT INTO cards (id, nid, did, ord, mod, usn, type, queue, due, ivl, factor, reps, lapses, left, odue, odid, flags, data)
				VALUES (?, ?, ?, ?, ?, -1, 0, 0, ?, 0, 0, 0, 0, 0, 0, 0, 0, '')
			`, cardID, noteID, deckID, ord, now, i+1)
			if err != nil {
				return fmt.Errorf("failed to insert card: %w", err)
			}
		}
	}

	return nil
}

func fieldChecksum(s string) int64 {
	var sum int64
	for _, r := range s {
		sum = (sum*31 + int64(r)) % 2147483647
	}
	return sum
}

func createAPKG(tempDir, outputPath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	// Add collection.anki2
	dbPath := filepath.Join(tempDir, "collection.anki2")
	dbContent, err := os.ReadFile(dbPath)
	if err != nil {
		return err
	}

	dbWriter, err := zipWriter.Create("collection.anki2")
	if err != nil {
		return err
	}
	if _, err := dbWriter.Write(dbContent); err != nil {
		return err
	}

	// Add media file (empty JSON object for no media)
	mediaWriter, err := zipWriter.Create("media")
	if err != nil {
		return err
	}
	if _, err := mediaWriter.Write([]byte("{}")); err != nil {
		return err
	}

	return nil
}

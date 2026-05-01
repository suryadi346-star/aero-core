package cache

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"github.com/suryadi346-star/aero-core/internal/session"
)

type SQLiteCache struct { db *sql.DB }

func NewSQLiteCache(db *sql.DB) (*SQLiteCache, error) {
	c := &SQLiteCache{db: db}
	_, err := c.db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (id TEXT PRIMARY KEY, data TEXT NOT NULL, updated_at INTEGER NOT NULL);
		CREATE TABLE IF NOT EXISTS response_cache (prompt_hash TEXT PRIMARY KEY, response TEXT NOT NULL, created_at INTEGER NOT NULL);
	`)
	return c, err
}

func (c *SQLiteCache) SaveSession(sess *session.Session) error {
	data, _ := json.Marshal(sess)
	_, err := c.db.Exec(`INSERT OR REPLACE INTO sessions (id, data, updated_at) VALUES (?, ?, ?)`, sess.ID, string(data), sess.Updated.Unix())
	return err
}

func (c *SQLiteCache) LoadSession(id string) (*session.Session, error) {
	var data string
	err := c.db.QueryRow(`SELECT data FROM sessions WHERE id = ?`, id).Scan(&data)
	if err == sql.ErrNoRows { return nil, nil }
	if err != nil { return nil, err }
	var sess session.Session
	json.Unmarshal([]byte(data), &sess)
	return &sess, nil
}

func PromptHash(system, user string) string {
	h := sha256.Sum256([]byte(system + "||" + user))
	return fmt.Sprintf("%x", h)
}

func (c *SQLiteCache) GetCachedResponse(hash string) (string, error) {
	var resp string
	err := c.db.QueryRow(`SELECT response FROM response_cache WHERE prompt_hash = ?`, hash).Scan(&resp)
	if err == sql.ErrNoRows { return "", nil }
	return resp, err
}

func (c *SQLiteCache) CacheResponse(hash, response string) error {
	_, err := c.db.Exec(`INSERT OR REPLACE INTO response_cache (prompt_hash, response, created_at) VALUES (?, ?, ?)`, hash, response, time.Now().Unix())
	return err
}

func (c *SQLiteCache) PruneCache() error {
	_, err := c.db.Exec(`DELETE FROM response_cache WHERE rowid NOT IN (SELECT rowid FROM response_cache ORDER BY created_at DESC LIMIT 1000)`)
	return err
}

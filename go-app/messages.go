package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Messages []Message

type Message struct {
	Id        int       `json:"id"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	UserId    int       `json:"user_id"`
	UserName  string    `json:"user_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func initMessagesEndpoints(router *gin.Engine) {
	router.POST("/messages", func(c *gin.Context) {
		authUserId, err := strconv.Atoi(c.GetHeader("X-User-Id"))
		if err != nil || authUserId == 0 {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}

		message := Message{}

		if err := c.ShouldBind(&message); err != nil {
			errResponse(c, err)
			return
		}

		message.UserId = authUserId
		message.UserName = c.GetHeader("X-User-Name")

		stat, err := createMessage(&message)
		if err != nil {
			errResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, stat)
	})

	if viper.GetBool("cache-enabled") {
		router.GET("/messages", cachePageIfValid(cacheStore, 10 * time.Minute, messagesUnchanged, readMessagesFunc()))
		router.GET("/messages/:id", cache.CachePage(cacheStore, 10 * time.Minute, readMessageFunc()))
	} else {
		router.GET("/messages", readMessagesFunc())
		router.GET("/messages/:id", readMessageFunc())
	}

	router.PUT("/messages/:id", func(c *gin.Context) {
		authUserId, err := strconv.Atoi(c.GetHeader("X-User-Id"))
		if err != nil || authUserId == 0 {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}

		message, err := getMessageById(c.Param("id"))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{})
			} else {
				errResponse(c, err)
			}
			return
		}
		if err := c.ShouldBind(message); err != nil {
			errResponse(c, err)
			return
		}

		dbMessage, err := getMessageById(c.Param("id"))
		if err != nil {
			errResponse(c, err)
			return
		}
		if dbMessage.UserId != authUserId {
			c.JSON(http.StatusUnauthorized, "Not my message")
			return
		}

		stat, err := updateMessageById(c.Param("id"), message)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, stat)
	})

	router.DELETE("/messages/:id", func(c *gin.Context) {
		authUserId, err := strconv.Atoi(c.GetHeader("X-User-Id"))
		if err != nil || authUserId == 0 {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}

		dbMessage, err := getMessageById(c.Param("id"))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{})
			} else {
				errResponse(c, err)
			}
			return
		}

		if dbMessage.UserId != authUserId {
			c.JSON(http.StatusUnauthorized, "Not my message")
			return
		}

		stat, err := deleteMessageById(c.Param("id"))
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, stat)
	})
}

func messagesUnchanged() bool {
	var lastUpdatedCached int64
	key := "messages:updated"
	lastUpdated, _ := getLastMessagesUpdatedAt()
	if !lastUpdated.Valid {
		return true
	}

	if err := cacheStore.Get(key, &lastUpdatedCached); err != nil {
		if err != persistence.ErrCacheMiss {
			log.Println(err.Error())
		}
		cacheStore.Add(key, lastUpdated.Time.Unix(), CacheLifeATime)
		return false
	}

	if lastUpdated.Time.Unix() > lastUpdatedCached {
		cacheStore.Replace(key, lastUpdated.Time.Unix(), CacheLifeATime)
		return false
	}
	return true
}

func readMessagesFunc() func(c *gin.Context) {
	return func(c *gin.Context) {
		var err error
		limit := 10
		offset := 0
		search := ""

		if len(c.Query("limit")) > 0 {
			limit, err = strconv.Atoi(c.Query("limit"))
			if err != nil {
				errResponse(c, err)
				return
			}
		}
		if len(c.Query("offset")) > 0 {
			offset, err = strconv.Atoi(c.Query("offset"))
			if err != nil {
				errResponse(c, err)
				return
			}
		}
		if len(c.Query("search")) > 0 {
			search = c.Query("search")
		}

		messages, err := getMessages(limit, offset, search)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, messages)
	}
}

func readMessageFunc() func(c *gin.Context) {
	return func(c *gin.Context) {
		message, err := getMessageById(c.Param("id"))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{})
			} else {
				errResponse(c, err)
			}
			return
		}

		c.JSON(http.StatusOK, message)
	}
}

func getLastMessagesUpdatedAt() (mysql.NullTime, error) {
	var updated mysql.NullTime

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return updated, err
	}

	dbRow := db.QueryRow("SELECT updated_at FROM messages ORDER BY id DESC LIMIT 1")
	err = dbRow.Scan(&updated)
	if err != nil {
		return updated, err
	}

	return updated, nil
}

func getMessages(limit, offset int, search string) (*Messages, error) {
	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	var selDB *sql.Rows
	if len(search) > 0 {
		selDB, err = db.Query("SELECT id, subject, body, user_id, user_name, created_at, updated_at FROM messages WHERE body like ? ORDER BY created_at DESC LIMIT ? OFFSET ?", "%" + search + "%", limit, offset)
	} else {
		selDB, err = db.Query("SELECT id, subject, body, user_id, user_name, created_at, updated_at FROM messages ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	}
	if err != nil {
		return nil, err
	}
	message := Message{}
	messages := Messages{}

	for selDB.Next() {
		err = fetchDBMessage(&message, selDB)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return &messages, nil
}

func getMessageById(id string) (*Message, error) {
	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return nil, err
	}
	dbRow := db.QueryRow("SELECT id, subject, body, user_id, user_name, created_at, updated_at FROM messages WHERE id = ?", id)
	message := &Message{}
	err = fetchDBMessage(message, dbRow)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func createMessage(newData *Message) (ExecStats, error) {
	res := ExecStats{}

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return res, err
	}
	createStatement, err := db.Prepare("INSERT INTO messages(subject, body, user_id, user_name, created_at, updated_at) VALUES(?,?,?,?,?,?)")
	if err != nil {
		return res, err
	}
	created, err := createStatement.Exec(newData.Subject, newData.Body, newData.UserId, newData.UserName, time.Now(), time.Now())
	if err != nil {
		return res, err
	}
	affected, err := created.RowsAffected()
	if err != nil {
		return res, err
	}
	newId, err := created.LastInsertId()
	if err != nil {
		return res, err
	}

	res.Affected = affected
	res.LastInsertId = newId
	return res, nil
}

func updateMessageById(id string, newData *Message) (ExecStats, error) {
	res := ExecStats{}

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return res, err
	}
	updateStatement, err := db.Prepare("UPDATE messages SET subject=?,body=?,updated_at=? WHERE id=?")
	if err != nil {
		return res, err
	}
	created, err := updateStatement.Exec(newData.Subject, newData.Body, time.Now(), id)
	if err != nil {
		return res, err
	}
	affected, err := created.RowsAffected()
	if err != nil {
		return res, err
	}
	res.Affected = affected
	return res, nil
}

func deleteMessageById(id string) (ExecStats, error) {
	res := ExecStats{}

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return res, err
	}
	delStatement, err := db.Prepare("DELETE FROM messages WHERE id = ?")
	if err != nil {
		return res, err
	}
	deleted, err := delStatement.Exec(id)
	if err != nil {
		return res, err
	}
	affected, err := deleted.RowsAffected()
	if err != nil {
		return res, err
	}
	res.Affected = affected
	return res, nil
}

func fetchDBMessage(message *Message, row interface{}) error {
	var created, updated mysql.NullTime
	var err error
	switch rowValue := row.(type) {
	case *sql.Rows:
		err = rowValue.Scan(&message.Id, &message.Subject, &message.Body, &message.UserId, &message.UserName, &created, &updated)
	case *sql.Row:
		err = rowValue.Scan(&message.Id, &message.Subject, &message.Body, &message.UserId, &message.UserName, &created, &updated)
	default:
		return fmt.Errorf("unknow type for raw: %v", row)
	}
	if err != nil {
		return err
	}

	if created.Valid {
		message.CreatedAt = created.Time
	}
	if updated.Valid {
		message.UpdatedAt = updated.Time
	} else {
		message.UpdatedAt = message.CreatedAt
	}
	return nil
}

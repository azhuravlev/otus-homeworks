package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
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
	router.GET("/messages", func(c *gin.Context) {
		var err error
		limit := 10
		offset := 0

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

		messages, err := getMessages(limit, offset)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, messages)
	})

	router.POST("/messages", func(c *gin.Context) {
		message := Message{}
		if err := c.ShouldBind(&message); err != nil {
			errResponse(c, err)
			return
		}
		stat, err := createMessage(&message)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, stat)
	})

	router.GET("/messages/:id", func(c *gin.Context) {
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
	})

	router.PUT("/messages/:id", func(c *gin.Context) {
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
		stat, err := updateMessageById(c.Param("id"), message)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, stat)
	})

	router.DELETE("/messages/:id", func(c *gin.Context) {
		_, err := getMessageById(c.Param("id"))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{})
			} else {
				errResponse(c, err)
			}
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

func getMessages(limit, offset int) (*Messages, error) {
	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	selDB, err := db.Query("SELECT id, subject, body, user_id, user_name, created_at, updated_at FROM messages ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
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

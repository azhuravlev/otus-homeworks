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

type Users []User

type User struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func initUsersEndpoints(router *gin.Engine) {
	router.GET("/users", func(c *gin.Context) {
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

		users, err := getUsers(limit, offset)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, users)
	})

	router.POST("/users", func(c *gin.Context) {
		user := User{}
		if err := c.ShouldBind(&user); err != nil {
			errResponse(c, err)
			return
		}
		stat, err := createUser(&user)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, stat)
	})

	router.GET("/users/:id", func(c *gin.Context) {
		user, err := getUserById(c.Param("id"))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{})
			} else {
				errResponse(c, err)
			}
			return
		}

		c.JSON(http.StatusOK, user)
	})

	router.PUT("/users/:id", func(c *gin.Context) {
		user, err := getUserById(c.Param("id"))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{})
			} else {
				errResponse(c, err)
			}
			return
		}
		if err := c.ShouldBind(user); err != nil {
			errResponse(c, err)
			return
		}
		stat, err := updateUserById(c.Param("id"), user)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, stat)
	})

	router.DELETE("/users/:id", func(c *gin.Context) {
		_, err := getUserById(c.Param("id"))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{})
			} else {
				errResponse(c, err)
			}
			return
		}

		stat, err := deleteUserById(c.Param("id"))
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, stat)
	})
}

func getUsers(limit, offset int) (*Users, error) {
	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	selDB, err := db.Query("SELECT id, name, email, created_at, updated_at FROM users ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	user := User{}
	users := Users{}

	for selDB.Next() {
		err = fetchDBUser(&user, selDB)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return &users, nil
}

func getUserById(userId string) (*User, error) {
	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return nil, err
	}
	dbRow := db.QueryRow("SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?", userId)
	user := &User{}
	err = fetchDBUser(user, dbRow)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func createUser(newData *User) (ExecStats, error) {
	res := ExecStats{}

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return res, err
	}
	createStatement, err := db.Prepare("INSERT INTO users(name, email, created_at, updated_at ) VALUES(?,?,?,?)")
	if err != nil {
		return res, err
	}
	created, err := createStatement.Exec(newData.Name, newData.Email, time.Now(), time.Now())
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

func updateUserById(userId string, newData *User) (ExecStats, error) {
	res := ExecStats{}

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return res, err
	}
	updateStatement, err := db.Prepare("UPDATE users SET name=?,email=?,updated_at=? WHERE id=?")
	if err != nil {
		return res, err
	}
	created, err := updateStatement.Exec(newData.Name, newData.Email, time.Now(), userId)
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

func deleteUserById(userId string) (ExecStats, error) {
	res := ExecStats{}

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return res, err
	}
	delStatement, err := db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		return res, err
	}
	deleted, err := delStatement.Exec(userId)
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

func fetchDBUser(user *User, row interface{}) error {
	var created, updated mysql.NullTime
	var err error
	switch rowValue := row.(type) {
	case *sql.Rows:
		err = rowValue.Scan(&user.Id, &user.Name, &user.Email, &created, &updated)
	case *sql.Row:
		err = rowValue.Scan(&user.Id, &user.Name, &user.Email, &created, &updated)
	default:
		return fmt.Errorf("unknow type for raw: %v", row)
	}
	if err != nil {
		return err
	}

	if created.Valid {
		user.CreatedAt = created.Time
	}
	if updated.Valid {
		user.UpdatedAt = updated.Time
	} else {
		user.UpdatedAt = user.CreatedAt
	}
	return nil
}

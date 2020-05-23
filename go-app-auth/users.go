package main

import (
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

type User struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

const JWTLifeTime = time.Minute * 5

func initUsersEndpoints(router *gin.Engine) {
	router.POST("/register", func(c *gin.Context) {
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

	router.POST("/login", func(c *gin.Context) {
		user := &User{}
		if err := c.ShouldBind(user); err != nil {
			errResponse(c, err)
			return
		}

		dbUser, err := getUserByEmail(user.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{})
			} else {
				errResponse(c, err)
			}
			return
		}

		if user.Password != dbUser.Password {
			c.JSON(http.StatusUnauthorized, "Please provide valid login details")
			return
		}

		token, err := createToken(dbUser)

		if err != nil {
			errResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, token)
	})
}

func getUserByEmail(userEmail string) (*User, error) {
	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return nil, err
	}
	dbRow := db.QueryRow("SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = ?", userEmail)
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
	createStatement, err := db.Prepare("INSERT INTO users(name, email, password, created_at, updated_at ) VALUES(?,?,?,?,?)")
	if err != nil {
		return res, err
	}
	created, err := createStatement.Exec(newData.Name, newData.Email, newData.Password, time.Now(), time.Now())
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

func fetchDBUser(user *User, row interface{}) error {
	var created, updated mysql.NullTime
	var err error
	switch rowValue := row.(type) {
	case *sql.Rows:
		err = rowValue.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &created, &updated)
	case *sql.Row:
		err = rowValue.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &created, &updated)
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

func createToken(user *User) (string, error) {
	var err error
	secret := viper.GetString("secret")

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = user.Id
	atClaims["user_name"] = user.Name
	atClaims["user_email"] = user.Email
	atClaims["exp"] = time.Now().Add(JWTLifeTime).Unix()

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(secret))

	if err != nil {
		return "", err
	}

	return token, nil
}

package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Orders []Order

type Order struct {
	Id        int       `json:"id"`
	ProductId int       `json:"product_id"`
	Product   *Product  `json:"product"`
	Count     int       `json:"count"`
	Total     int       `json:"total"`
	UserId    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func initOrdersEndpoints(router *gin.Engine) {
	router.POST("/orders", func(c *gin.Context) {
		authUserId, err := strconv.Atoi(c.GetHeader("X-User-Id"))
		if err != nil || authUserId == 0 {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}

		etag := strconv.Itoa(getLastOrderId(authUserId))

		if match := c.GetHeader("If-Match"); match != "" {
			if !strings.Contains(match, etag) {
				c.JSON(http.StatusConflict, "Orders was changed")
				return
			}
		}

		order := Order{}

		if err := c.ShouldBind(&order); err != nil {
			errResponse(c, err)
			return
		}

		order.UserId = authUserId

		stat, err := createOrder(&order)
		if err != nil {
			errResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, stat)
	})
	router.GET("/orders", readOrdersFunc())
	router.GET("/orders/:id", readOrderFunc())
	router.DELETE("/orders/:id", func(c *gin.Context) {
		authUserId, err := strconv.Atoi(c.GetHeader("X-User-Id"))
		if err != nil || authUserId == 0 {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}

		dbOrder, err := getOrderById(c.Param("id"), authUserId)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{})
			} else {
				errResponse(c, err)
			}
			return
		}

		if dbOrder.UserId != authUserId {
			c.JSON(http.StatusUnauthorized, "Not my order")
			return
		}

		stat, err := deleteOrderById(c.Param("id"))
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, stat)
	})
}

func readOrdersFunc() func(c *gin.Context) {
	return func(c *gin.Context) {
		var err error

		authUserId, err := strconv.Atoi(c.GetHeader("X-User-Id"))
		if err != nil {
			authUserId = 0
		}

		orders, err := getOrders(authUserId)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.Header("ETag", strconv.Itoa(getLastOrderId(authUserId)))
		c.JSON(http.StatusOK, orders)
	}
}

func readOrderFunc() func(c *gin.Context) {
	return func(c *gin.Context) {
		authUserId, err := strconv.Atoi(c.GetHeader("X-User-Id"))
		if err != nil {
			authUserId = 0
		}

		order, err := getOrderById(c.Param("id"), authUserId)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{})
			} else {
				errResponse(c, err)
			}
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

func getLastOrderId(user_id int) int {
	var lastId int

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return lastId
	}

	dbRow := db.QueryRow("SELECT id FROM orders WHERE user_id = ? ORDER BY id DESC LIMIT 1", user_id)
	err = dbRow.Scan(&lastId)
	if err != nil {
		return lastId
	}

	return lastId
}

func getOrders(user_id int) (*Orders, error) {
	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	var selDB *sql.Rows
	selDB, err = db.Query("SELECT id, product_id, count, user_id, created_at, updated_at FROM orders WHERE user_id = ? ORDER BY created_at DESC", user_id)
	if err != nil {
		return nil, err
	}
	order := Order{}
	orders := Orders{}

	for selDB.Next() {
		err = fetchDBOrder(&order, selDB)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return &orders, nil
}

func getOrderById(id string, user_id int) (*Order, error) {
	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return nil, err
	}
	dbRow := db.QueryRow("SELECT id, product_id, count, user_id, created_at, updated_at FROM orders WHERE id = ? AND user_id = ?", id, user_id)
	order := &Order{}
	err = fetchDBOrder(order, dbRow)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func createOrder(newData *Order) (ExecStats, error) {
	res := ExecStats{}

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return res, err
	}
	createStatement, err := db.Prepare("INSERT INTO orders(product_id, count, user_id) VALUES(?,?,?)")
	if err != nil {
		return res, err
	}
	created, err := createStatement.Exec(newData.ProductId, newData.Count, newData.UserId)
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

func deleteOrderById(id string) (ExecStats, error) {
	res := ExecStats{}

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return res, err
	}
	delStatement, err := db.Prepare("DELETE FROM orders WHERE id = ?")
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

func fetchDBOrder(order *Order, row interface{}) error {
	var created, updated mysql.NullTime
	var err error
	switch rowValue := row.(type) {
	case *sql.Rows:
		err = rowValue.Scan(&order.Id, &order.ProductId, &order.Count, &order.UserId, &created, &updated)
	case *sql.Row:
		err = rowValue.Scan(&order.Id, &order.ProductId, &order.Count, &order.UserId, &created, &updated)
	default:
		return fmt.Errorf("unknow type for raw: %v", row)
	}
	if err != nil {
		return err
	}

	if created.Valid {
		order.CreatedAt = created.Time
	}
	if updated.Valid {
		order.UpdatedAt = updated.Time
	} else {
		order.UpdatedAt = order.CreatedAt
	}

	if order.ProductId > 0 {
		product, err := getProductById(strconv.Itoa(order.ProductId))
		if err != nil {
			return nil
		}
		order.Product = product
		order.Total = order.Count * product.Price
	}
	return nil
}

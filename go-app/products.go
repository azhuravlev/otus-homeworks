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

type Products []Product

type Product struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Price     int       `json:"price"`
	Available int       `json:"available"`
	Reserved  int       `json:"reserved"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func initProductsEndpoints(router *gin.Engine) {
	router.GET("/products", readProductsFunc())
	router.GET("/products/:id", readProductFunc())
	router.POST("/products", func(c *gin.Context) {
		admin := c.GetHeader("X-User-Admin")
		if admin == "" {
			c.JSON(http.StatusUnprocessableEntity, "Not allowed")
			return
		}

		product := Product{}

		if err := c.ShouldBind(&product); err != nil {
			errResponse(c, err)
			return
		}

		stat, err := createProduct(&product)
		if err != nil {
			errResponse(c, err)
			return
		}

		c.JSON(http.StatusOK, stat)
	})
	router.PUT("/products/:id", func(c *gin.Context) {
		admin := c.GetHeader("X-User-Admin")
		if admin == "" {
			c.JSON(http.StatusUnprocessableEntity, "Not allowed")
			return
		}

		product, err := getProductById(c.Param("id"))
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{})
		} else {
			errResponse(c, err)
		}

		stat, err := updateProductById(c.Param("id"), product)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, stat)
	})
	router.PUT("/products/:id/reserve", func(c *gin.Context) {
		admin := c.GetHeader("X-User-Admin")
		if admin == "" {
			c.JSON(http.StatusUnprocessableEntity, "Not allowed")
			return
		}

		product, err := getProductById(c.Param("id"))
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{})
		} else {
			errResponse(c, err)
		}

		count, err := strconv.Atoi(c.Param("count"))
		if err != nil {
			errResponse(c, err)
			return
		}
		err = product.Reserve(count)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, "Reserved")
	})
	router.PUT("/products/:id/unreserve", func(c *gin.Context) {
		admin := c.GetHeader("X-User-Admin")
		if admin == "" {
			c.JSON(http.StatusUnprocessableEntity, "Not allowed")
			return
		}

		product, err := getProductById(c.Param("id"))
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{})
		} else {
			errResponse(c, err)
		}

		count, err := strconv.Atoi(c.Param("count"))
		if err != nil {
			errResponse(c, err)
			return
		}
		err = product.ReturnFromReserve(count)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, "Unreserved")
	})
	router.PUT("/products/:id/sell", func(c *gin.Context) {
		admin := c.GetHeader("X-User-Admin")
		if admin == "" {
			c.JSON(http.StatusUnprocessableEntity, "Not allowed")
			return
		}

		product, err := getProductById(c.Param("id"))
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{})
		} else {
			errResponse(c, err)
		}

		count, err := strconv.Atoi(c.Param("count"))
		if err != nil {
			errResponse(c, err)
			return
		}
		err = product.Sell(count)
		if err != nil {
			errResponse(c, err)
			return
		}
		c.JSON(http.StatusOK, "Sold")
	})
}

func readProductsFunc() func(c *gin.Context) {
	return func(c *gin.Context) {
		var err error
		products, err := getProducts()
		if err != nil {
			errResponse(c, err)
			return
		}
		c.Header("ETag", strconv.Itoa(getLastProductId()))
		c.JSON(http.StatusOK, products)
	}
}

func readProductFunc() func(c *gin.Context) {
	return func(c *gin.Context) {
		product, err := getProductById(c.Param("id"))
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{})
			} else {
				errResponse(c, err)
			}
			return
		}

		c.JSON(http.StatusOK, product)
	}
}

func getLastProductId() int {
	var lastId int

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return lastId
	}

	dbRow := db.QueryRow("SELECT id FROM products order BY id DESC LIMIT 1")
	err = dbRow.Scan(&lastId)
	if err != nil {
		return lastId
	}

	return lastId
}

func getProducts() (*Products, error) {
	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return nil, err
	}

	var selDB *sql.Rows
	selDB, err = db.Query("SELECT id, name, price, available, reserved, created_at, updated_at FROM products order BY created_at DESC")
	if err != nil {
		return nil, err
	}
	product := Product{}
	products := Products{}

	for selDB.Next() {
		err = fetchDBProduct(&product, selDB)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return &products, nil
}

func getProductById(id string) (*Product, error) {
	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return nil, err
	}
	dbRow := db.QueryRow("SELECT id, name, price, available, reserved, created_at, updated_at FROM products WHERE id = ?", id)
	product := &Product{}
	err = fetchDBProduct(product, dbRow)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func fetchDBProduct(product *Product, row interface{}) error {
	var created, updated mysql.NullTime
	var err error
	switch rowValue := row.(type) {
	case *sql.Rows:
		err = rowValue.Scan(&product.Id, &product.Name, &product.Price, &product.Available, &product.Reserved, &created, &updated)
	case *sql.Row:
		err = rowValue.Scan(&product.Id, &product.Name, &product.Price, &product.Available, &product.Reserved, &created, &updated)
	default:
		return fmt.Errorf("unknow type for raw: %v", row)
	}
	if err != nil {
		return err
	}

	if created.Valid {
		product.CreatedAt = created.Time
	}
	if updated.Valid {
		product.UpdatedAt = updated.Time
	} else {
		product.UpdatedAt = product.CreatedAt
	}
	return nil
}

func createProduct(newData *Product) (ExecStats, error) {
	res := ExecStats{}

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return res, err
	}
	createStatement, err := db.Prepare("INSERT INTO products(name, price, available) VALUES(?,?,?)")
	if err != nil {
		return res, err
	}
	created, err := createStatement.Exec(newData.Name, newData.Price, newData.Available)
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

func updateProductById(id string, newData *Product) (ExecStats, error) {
	res := ExecStats{}

	db, err := dbConn()
	defer db.Close()
	if err != nil {
		return res, err
	}
	updateStatement, err := db.Prepare("UPDATE products SET name=?,price=?,available=?,reserved=? WHERE id=?")
	if err != nil {
		return res, err
	}
	created, err := updateStatement.Exec(newData.Name, newData.Price, newData.Available, newData.Reserved, id)
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

func (product *Product) Reserve(count int) error {
	product.Reserved = product.Reserved + count
	if product.Reserved > product.Available {
		return fmt.Errorf("Not enough products available only: %d", product.Available)
	}

	_, err := updateProductById(strconv.Itoa(product.Id), product)
	return err
}

func (product *Product) ReturnFromReserve(count int) error {
	product.Reserved = product.Reserved - count
	_, err := updateProductById(strconv.Itoa(product.Id), product)
	return err
}

func (product *Product) Sell(count int) error {
	if product.Reserved > product.Available {
		return fmt.Errorf("Not enough products available only: %d", product.Available)
	}
	product.Reserved = product.Reserved - count
	product.Available = product.Available - count

	_, err := updateProductById(strconv.Itoa(product.Id), product)
	return err
}

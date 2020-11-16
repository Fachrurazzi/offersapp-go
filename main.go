package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/RazzyTirta/offersapp/models"
	"github.com/RazzyTirta/offersapp/routes"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"golang.org/x/net/context"
)

func main() {

	conn, err := connectDB()
	if err != nil {
		return
	}

	router := gin.Default() 

	router.Use(dbMiddleware(*conn))

	usersGroup := router.Group("users")
	{
		usersGroup.POST("register", routes.UserRegister)
		usersGroup.POST("login", routes.UserLogin)
	}

	itemsGroup := router.Group("items")
	{
		itemsGroup.GET("index", routes.ItemsIndex)
		itemsGroup.POST("create", authMiddleware(), routes.ItemsCreate)
		itemsGroup.GET("sold_by_user", authMiddleware(), routes.ItemsForSaleByCurrentUser)
	}

	router.Run(":3000")
}

func connectDB() (c *pgx.Conn, err error) {
	conn, err := pgx.Connect(context.Background(), "postgresql://postgres:postgres@localhost:5432/offersapp")
	if err != nil {
		fmt.Println("Error connecting to DB")
		panic(err.Error())
	}

	_ = conn.Ping(context.Background())

	return conn, err
}

func dbMiddleware(conn pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context)  {
		c.Set("db", conn)
		c.Next()

	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearer := c.Request.Header.Get("Authorization")
		split := strings.Split(bearer, "Bearer ")
		if len(split) < 2 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Not Authenticated.",
			})
			c.Abort()
			return
		}
		token := split[1]
		isValid, userID := models.IsTokenValid(token)
		if isValid == false {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Not Authenticated.",
			})
			c.Abort()
		} else {
			c.Set("user_id", userID)
			c.Next()
		}
	}
}
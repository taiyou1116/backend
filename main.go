package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// DB情報
	const (
		host     = "db"
		port     = 5432
		user     = "user"
		password = "password"
		dbname   = "mydatabase"
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected!")

	e := gin.Default()
	e.Use(cors.Default())

	e.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Hello World",
		})
	})

	// ここでusersテーブルからデータを取得するルートを追加
	e.GET("/users", func(c *gin.Context) {

		rows, err := db.Query("SELECT id, username, email FROM users")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var users []gin.H
		for rows.Next() {
			var id int
			var username, email string
			err = rows.Scan(&id, &username, &email)
			if err != nil {
				log.Fatal(err)
			}
			users = append(users, gin.H{"id": id, "username": username, "email": email})
		}

		c.JSON(200, gin.H{"users": users})
	})

	e.Run(":8000")
}

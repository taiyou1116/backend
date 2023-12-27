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
	e.GET("/posts", func(c *gin.Context) {

		// rows, err := db.Query("SELECT id, user_id, title, body FROM posts")
		rows, err := db.Query(`
        	SELECT posts.id, posts.user_id, posts.title, posts.body, users.username 
        	FROM posts 
        	INNER JOIN users ON posts.user_id = users.id
    	`)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var posts []gin.H

		for rows.Next() {
			var id, user_id int
			var title, body, username string
			err = rows.Scan(&id, &user_id, &title, &body, &username)
			if err != nil {
				log.Fatal(err)
			}
			posts = append(posts, gin.H{"id": id, "user_id": user_id, "title": title, "body": body, "username": username})
		}

		c.JSON(200, gin.H{"posts": posts})
	})

	e.Run(":8000")
}

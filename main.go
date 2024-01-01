package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
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
	// ミドルウェア
	e.Use(cors.Default())

	// 全postsを取得 (+username)
	// 10個ずつ取得とかに変更するかも
	e.GET("api/posts", func(c *gin.Context) {
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

		// 投稿の数だけループ
		for rows.Next() {
			var id, user_id int
			var title, body, username string
			// 送信された型と同じか確かめる
			err = rows.Scan(&id, &user_id, &title, &body, &username)
			if err != nil {
				log.Fatal(err)
			}
			posts = append(posts, gin.H{"id": id, "user_id": user_id, "title": title, "body": body, "username": username})
		}

		c.JSON(200, gin.H{"posts": posts})
	})

	// postをDBに追加
	e.POST("api/submit-post", func(c *gin.Context) {
		var payload PostPayload
		// payloadの型に合わせてバインド
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec("INSERT INTO posts (title, body, user_id) VALUES ($1, $2, $3)", payload.Title, payload.Content, payload.UserId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "created new post!"})
	})

	// ユーザー登録
	e.POST("api/create-user", func(c *gin.Context) {
		var payload UserPayload
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// パスワードをハッシュ化
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		_, err = db.Exec("INSERT INTO users (username, password) VALUE ($1, $2)", payload.UserName, hashedPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "created new User!"})
	})
	e.Run(":8000")
}

// POSTをマッピングするための構造体
type PostPayload struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	UserId  int    `json:"user_id"`
}

type UserPayload struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
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
	e.POST("api/register", func(c *gin.Context) {
		// クライアントからの値をバインド
		var payload UserPayload
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var userCount byte
		// ユーザーがすでに存在しているか確認
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", payload.UserName).Scan(&userCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if userCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "そのユーザー名はすでに使用されています"})
			return
		}

		// パスワードをハッシュ化
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// ユーザーをDBに追加
		_, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", payload.UserName, string(hashedPassword))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "created new User!"})
	})

	// ログイン処理
	e.POST("api/login", func(c *gin.Context) {
		// クライアントからの値をバインド
		var payload UserPayload
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		var storedHash string

		// passwordを入手してScan()でstoredHashに格納する
		err := db.QueryRow("SELECT password FROM users WHERE username = $1", payload.UserName).Scan(&storedHash)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザーが存在しません"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		// passwordの確認
		err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(payload.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "パスワードが違います"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ログイン成功"})

		// JWTトークンの生成
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": payload.UserName,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})
		tokenString, err := token.SignedString([]byte("your_secret_key"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "トークンの生成に失敗しました"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	})

	e.Run(":8000")
}

// StatusBadRequest          ... クライアントからのリクエストに問題あり
// StatusInternalServerError ... リクエストに問題なし、サーバーの処理で問題あり

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

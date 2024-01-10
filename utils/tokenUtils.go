package utils

import (
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// **他のパッケージから取得する場合は大文字から**
func VerifyToken(c *gin.Context) (string, error) {
	tokenString := c.GetHeader("Authorization")

	// 'Bearer 'プレフィックスを取り除く
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// トークン検証
	// tokenString -> 無名関数の引数のtokenに渡される
	// returnの結果がtokenに入る
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// ここで秘密鍵を返す
		return []byte("your_secret_key"), nil
	})

	if err != nil || !token.Valid {
		log.Printf("トークン解析エラー: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "無効なトークンです"})
		return "", nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "クレームを解説できません"})
		return "", nil
	}

	username := claims["username"].(string)

	// あとでレスポンス内容を変更する(DBから取得)
	c.JSON(http.StatusOK, gin.H{"response": username})

	return username, nil
}

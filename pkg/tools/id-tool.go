package tools

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/inooy/serco-client/pkg/snowflake"
	"io"
	"time"
)

var node *snowflake.SnowFlake

func init() {
	// Create a new Node with a Node number of 1
	sf, err := snowflake.New(1)
	if err != nil {
		panic(err)
	}
	node = sf
}

const SessionPrefix = "sess_"

func GetSnowflakeId() string {
	// Generate a snowflake ID.
	id, _ := node.Generate()
	return fmt.Sprintf("sess_map_%d", id)
}

func GetRandomToken(length int) string {
	r := make([]byte, length)
	io.ReadFull(rand.Reader, r)
	return base64.URLEncoding.EncodeToString(r)
}

func CreateSessionId(sessionId string) string {
	return SessionPrefix + sessionId
}

func GetSessionIdByUserId(userId int) string {
	return fmt.Sprintf("sess_map_%d", userId)
}

func GetSessionName(sessionId string) string {
	return SessionPrefix + sessionId
}

func Sha1(s string) (str string) {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func GetNowDateTime() string {
	return time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
}

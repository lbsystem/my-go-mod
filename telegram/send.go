package telegram

import (
	
	"net/http"
	"os"
)

func Send(str, token string) {
	chat_id := "606212656"
	url_req := "https://api.telegram.org/bot" + token + "/sendMessage" + "?chat_id=" + chat_id + "&text=" + str
	get, err := http.Get(url_req)
	if err != nil {
		return
	}
	os.Stdout.ReadFrom(get.Body)
}

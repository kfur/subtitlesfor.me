package controller

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/kfur/subtitler/app"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }} // use default options

func readLoop(c *websocket.Conn) {
	for {
		if _, _, err := c.NextReader(); err != nil {
			c.Close()
			break
		}
	}
}

func SRTHandler(w http.ResponseWriter, r *http.Request) {
	var closed = make(chan struct{})

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	c.SetCloseHandler(
		func(code int, text string) error {
			closed <- struct{}{}
			return nil
		})

	go readLoop(c)

	params := context.Get(r, "params").(httprouter.Params)

	token := params.ByName("token")
	tokenData := app.TokenMap[token]
	fmt.Printf("Received connection with token: %s\n", token)

	select {
	case subData := <-tokenData.Subtitles:

		err = c.WriteMessage(websocket.TextMessage, []byte(subData.SrtText))
		if err != nil {
			log.Println("write:", err)
		}

	case <-closed:
		fmt.Print("Tab closed prematurely. ")
		tokenData := app.TokenMap[token]
		if tokenData.Cancel != nil {
			(*(tokenData.Cancel))()
			fmt.Println("Cancel request")
		}
	}
	fmt.Printf("Will delete token: %s\n", token)

	delete(app.TokenMap, token)
}

package main

import (
	"encoding/json"
	"os"

	"html/template"

	"github.com/gorilla/websocket"

	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type M map[string]interface{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func homePage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	err2 := t.Execute(w, M{})
	if err2 != nil {
		fmt.Println("executing template:", err2)
	}

}

func reader(conn *websocket.Conn) {

	for {

		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		//fmt.Println(string(p))
		//fmt.Println(string(messageType))

		ai_answer := getTextFromAI(string(p))

		if err := conn.WriteMessage(messageType, []byte(ai_answer)); err != nil {
			fmt.Println(err)
			return
		}

	}

}

func wsEndPoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Printf("Client Succesfully Connected")

	first_res := getTextFromAI(`USER: Ask me personal question please \n WALTER:`)

	err = ws.WriteMessage(websocket.TextMessage, []byte(first_res))
	if err != nil {
		log.Println(err)
	}

	defer ws.Close()
	reader(ws)

}

func getTextFromAI(p string) string {
	content, err := os.ReadFile("walter.txt")
	if err != nil {
		fmt.Println(err)
	}

	client := http.Client{}
	payload := strings.NewReader(`{
		"model": "text-davinci-003",
		"prompt": "` + string(content) + `\n WALTER: ` + p + `",
		"temperature": 0.9,
		"max_tokens": 400,
		"top_p": 1,
		"frequency_penalty": 0.0,
		"presence_penalty": 0.6,
		"stop": [" WALTER:", " USER:"]
	  }`)

	//fmt.Println(payload)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/completions", payload)
	if err != nil {
		//Handle Error
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+"$API_KEY")

	res, err := client.Do(req)
	if err != nil {
		//Handle Error
		panic(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	//fmt.Println(string(body))

	var dat map[string]interface{}

	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}

	response_text := dat["choices"].([]interface{})[0].(map[string]interface{})["text"].(string)

	return response_text
}

func main() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndPoint)

	fs := http.FileServer(http.Dir("css"))

	http.Handle("/css/", http.StripPrefix("/css/", fs))

	log.Fatal(http.ListenAndServe(":80", nil))

}

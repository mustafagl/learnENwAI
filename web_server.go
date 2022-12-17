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

var chatHistory []string

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

		fmt.Println(string(p))
		fmt.Println(string(messageType))

		if err := conn.WriteMessage(messageType, p); err != nil {
			fmt.Println(err)
			return
		}
		chatHistory = append(chatHistory, "USER: "+string(p)+`\n WALTER:`)

		ai_answer := getTextFromAI()

		if err := conn.WriteMessage(messageType, []byte(ai_answer)); err != nil {
			fmt.Println(err)
			return
		}

		chatHistory = append(chatHistory, strings.ReplaceAll(ai_answer, "\n", ""))

	}

}

func wsEndPoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Client Succesfully Connected")
	chatHistory = nil
	chatHistory = append(chatHistory, `USER: Ask me personal question please \n WALTER:`)

	first_res := getTextFromAI()

	err = ws.WriteMessage(websocket.TextMessage, []byte(first_res))
	if err != nil {
		log.Println(err)
	}

	chatHistory = append(chatHistory, strings.ReplaceAll(first_res, "\n", ""))

	defer ws.Close()
	reader(ws)

}

func getTextFromAI() string {
	content, err := os.ReadFile("walter.txt")
	if err != nil {
		fmt.Println(err)
	}

	client := http.Client{}
	payload := strings.NewReader(`{
		"model": "text-davinci-003",
		"prompt": "` + string(content) + `\n` + strings.Join(chatHistory, `\n`) + `",
		"temperature": 0.9,
		"max_tokens": 400,
		"top_p": 1,
		"frequency_penalty": 0.0,
		"presence_penalty": 0.6,
		"stop": [" WALTER:", " USER:"]
	  }`)

	fmt.Println(payload)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/completions", payload)
	if err != nil {
		//Handle Error
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+"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")

	res, err := client.Do(req)
	if err != nil {
		//Handle Error
		panic(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	fmt.Println(string(body))

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

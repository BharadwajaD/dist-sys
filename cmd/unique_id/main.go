package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main(){
	node := maelstrom.NewNode()

	node.Handle("generate", func(msg maelstrom.Message) error {

		var body map[string]any
		err := json.Unmarshal(msg.Body, &body)
		if err != nil {
			return err
		}

		//unique id - timestamp+msg_id
		body["type"] = "generate_ok"
		t := time.Now()
		body["id"] = fmt.Sprintf("%d-%s",t.UnixNano(),msg.Src)

		return node.Reply(msg, body)
	})

	if err := node.Run(); err != nil{
		log.Fatal(err)
	}

}

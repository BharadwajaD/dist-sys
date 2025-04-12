package main

import (
    "encoding/json"
    "log"

    maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main(){
	node := maelstrom.NewNode()

	node.Handle("echo", func(msg maelstrom.Message) error {
		var req map[string]any
		err := json.Unmarshal(msg.Body, &req)
		if err != nil {
			return err
		}

		req["type"] = "echo_ok"
		return node.Reply(msg, req)
	})

	if err := node.Run(); err != nil{
		log.Fatal(err)
	}
}

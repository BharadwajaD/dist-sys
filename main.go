package main

import (
	"encoding/json"
	"log"
	"strconv"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func echo(n *maelstrom.Node){
	n.Handle("echo", func(msg maelstrom.Message) error {

        var body map[string]any
        err := json.Unmarshal(msg.Body, &body)

        if err != nil {
            return err
        }

        body["type"] = "echo_ok"
        body["in_reply_to"] = body["msg_id"]

		return n.Reply(msg, body)
	})

}


func main() {

	n := maelstrom.NewNode()
    n.Handle("generate", func(msg maelstrom.Message) error {

        var body map[string]any
        err := json.Unmarshal(msg.Body, &body)

        if err != nil {
            return err
        }

        body["type"] = "generate_ok"

        m_id := body["msg_id"].(float64)
        s := strconv.FormatFloat(m_id, 'f', -1, 64)
        body["in_reply_to"] = m_id
        body["id"] = msg.Dest + s

		return n.Reply(msg, body)
    })

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

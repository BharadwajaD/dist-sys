package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Content struct{
	Numbers []float64
	SentMessages map[string]float64
	SentMessageIds map[string]bool
	PendingRecievingNodes map[string][]string
}

func deleteFromList(list []string, ele string) []string {
	return slices.DeleteFunc(list, func(e string) bool { return e == ele})
}

func main(){
	node := maelstrom.NewNode()
	content := Content{
		SentMessageIds: make(map[string]bool),
		SentMessages: make(map[string]float64),
		Numbers: make([]float64, 0),
		PendingRecievingNodes: make(map[string][]string),
	}
	var adj_nodes []any 
	var mu sync.RWMutex

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//refresh loop
	go func(){
		//todo: graceful shutdown
		ticker := time.NewTicker(time.Duration(500*time.Millisecond))
		for {
			select {
			case <- ticker.C :
				for umsg_id , unsent_nodes := range content.PendingRecievingNodes {
					for _, nidx := range unsent_nodes {
						num := content.SentMessages[umsg_id]
						node.Send(nidx, map[string]any{"type": "broadcast", "message": num, "umsg_id": umsg_id})
					}
				}
			case <- ctx.Done(): return
			}
		}
	}()

	//TODO: Learn gossip protocols and implemnt them here...
	node.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		json.Unmarshal(msg.Body, &body)

		mu.Lock()
		defer mu.Unlock()
		num := body["message"].(float64)
		//new umsg_id bcoz msg_id is nt unique
		var umsg_id string
		if body["umsg_id"] == nil {
			umsg_id = fmt.Sprintf("%d-%s",time.Now().UnixNano(),msg.Src)
		}else{
			umsg_id = body["umsg_id"].(string)
		}

		var err error
		if content.SentMessageIds[umsg_id] == false {
			content.Numbers = append(content.Numbers, num)
			content.SentMessageIds[umsg_id] = true
			content.SentMessages[umsg_id] = num
			for _, nidx := range adj_nodes{
				node.Send(nidx.(string), map[string]any{"type": "broadcast", "message": num, "umsg_id": umsg_id})
				pending_nodes := content.PendingRecievingNodes[umsg_id]

				//todo: but here the sender might get killed  
				pending_nodes = append(pending_nodes, nidx.(string))
				content.PendingRecievingNodes[umsg_id] = pending_nodes
			}
		}
		if body["umsg_id"] == nil {
			err = node.Reply(msg, map[string]any{"type":"broadcast_ok"})
		}else {
			err = node.Reply(msg, map[string]any{"type":"broadcast_ok", "umsg_id": umsg_id})
		}

		return err
	})

	//var recieved_mu sync.Mutex
	node.Handle("broadcast_ok", func(msg maelstrom.Message) error {
		var body map[string]any
		json.Unmarshal(msg.Body, &body)

		umsg_id := body["umsg_id"].(string)
		//do i need mu here ??
		content.PendingRecievingNodes[umsg_id] = deleteFromList(content.PendingRecievingNodes[umsg_id], msg.Src)
		return nil
	})

	node.Handle("read", func(msg maelstrom.Message) error {
		mu.RLock()
		defer mu.RUnlock()
		return node.Reply(msg, map[string]any{"type":"read_ok", "messages": content.Numbers})
	})

	node.Handle("topology", func(msg maelstrom.Message) error {
		var body map[string]any
		json.Unmarshal(msg.Body, &body)

		top := body["topology"].(map[string]any)
		adj_nodes = top[node.ID()].([]any)
		return node.Reply(msg, map[string]any{"type":"topology_ok"})
	})

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}


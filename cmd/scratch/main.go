package main

import (
	"encoding/json"

	"github.com/0xor1/wtf/cmd/boring/pkg/blockers/blockerseps"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
)

func main() {
	testJsonSizeOfBlockersGame()
}

func testJsonSizeOfBlockersGame() {
	g1 := blockerseps.NewGame(app.ExampleID(), app.ExampleID())
	bs, err := json.Marshal(g1)
	Println(string(bs), len(bs), err)
}

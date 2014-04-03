// Copyright 2014 Johan Rydberg.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import "github.com/hashicorp/serf/client"
import "fmt"
import "encoding/json"
import "strings"
import "log"

type ICommand interface {
    Update(formation, identity string, payload JSONObject)
}

type SerfCommand struct {
    client *client.RPCClient
}

func NewSerfCommand(client *client.RPCClient) *SerfCommand {
    return &SerfCommand{client}
}

func (command *SerfCommand) Update(formation, identity string, payload JSONObject) {
    data, _ := json.Marshal(payload)
    body := fmt.Sprintf("%s:%s:%s", formation, identity, data)
    err := command.client.UserEvent("advertise", []byte(body), false)
    if err != nil {
        log.Printf("GOT ERROR ON SENDING: %s\n", err.Error())
    }
}

func (command *SerfCommand) Handle(registry IRegistry) {
    input := make(chan map[string]interface{}, 32)
    go command.client.Stream("", input)

    for {
        data := <-input
        name, ok := data["Name"].(string)
        if ok && name == "advertise" {
            parts := strings.SplitN(string(data["Payload"].([]byte)), ":", 3)

            var payload JSONObject
            err := json.Unmarshal([]byte(parts[2]), &payload)
            if err == nil {
                registry.Update(parts[0], parts[1], payload)
            }
        }
    }
}

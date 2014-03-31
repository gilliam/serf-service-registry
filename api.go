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

import "github.com/gorilla/mux"
import "net/http"
import "encoding/json"
import "io/ioutil"


type HttpApi struct {
    registry IRegistry
    command  ICommand
}

func NewHttpApi(registry IRegistry, command ICommand) *HttpApi {
    return &HttpApi{registry, command}
}

func (api *HttpApi) QueryFormation(w http.ResponseWriter, r *http.Request) {
    items := make(map[string]JSONObject)

    vars := mux.Vars(r)
    for k, entry := range api.registry.Index(vars["formation"]) {
        items[k] = entry.Payload
    }

    body, _ := json.Marshal(items)

    w.WriteHeader(http.StatusOK)
    w.Write(body)
}

func (api *HttpApi) Update(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    vars := mux.Vars(r)

    var payload JSONObject
    err = json.Unmarshal(body, &payload)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    api.command.Update(vars["formation"], vars["instance"], payload)
}

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
import "github.com/gorilla/mux"
import "github.com/gorilla/handlers"
import "net/http"
import "log"
import "os"


func main() {
    serf, err := client.NewRPCClient("127.0.0.1:7373")
    if err != nil {
        log.Fatal(err)
    }
    serfCommand := NewSerfCommand(serf)
    registry := NewRegistry()
    api := NewHttpApi(registry, serfCommand)

    router := mux.NewRouter()
    router.HandleFunc("/{formation}", api.QueryFormation).Methods("GET")
    router.HandleFunc("/{formation}/{instance}", api.Update).Methods("PUT")

    go serfCommand.Handle(registry)
    http.ListenAndServe(":3222", handlers.CombinedLoggingHandler(os.Stdout, router))
}

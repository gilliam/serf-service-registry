// X

package main

import "github.com/hashicorp/serf/client"
import "github.com/gorilla/mux"
import "net/http"
import "sync"
import "time"
import "fmt"
import "log"
import "encoding/json"
import "io/ioutil"

type JsonObject map[string]interface{}

type Instance struct {
    Formation string    `json: formation`
    Service string      `json: service`
    Instance string     `json: instance`
}

type Entry struct {
    payload JsonObject
    instance Instance
    timestamp time.Time
}

type Registry struct {
    sync.Mutex                    // lock for the registry
    entries map[string]Entry      // array of entries
    serf *client.RPCClient
}

func NewRegistry(serf *client.RPCClient) *Registry {
    return &Registry{sync.Mutex{}, make(map[string]Entry), serf}
}

func (reg *Registry) QueryFormationHandler(w http.ResponseWriter,
    r *http.Request) {
    var items map[string]JsonObject = make(map[string]JsonObject)

    vars := mux.Vars(r)
    formation := vars["formation"]

    reg.Lock()
    defer reg.Unlock()

    for k, entry := range reg.entries {
        if entry.instance.Formation == formation {
            items[k] = entry.payload
        }
    }

    body, _ := json.Marshal(items)

    w.WriteHeader(http.StatusOK)
    w.Write(body)
}

func (reg *Registry) UpdateHandler(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    var payload JsonObject
    err = json.Unmarshal(body, &payload)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    reg.serf.UserEvent("advertise", body, true)
}

func (reg *Registry) addEntryFromEvent(body []byte) {
    var inst Instance
    var payload JsonObject

    err := json.Unmarshal(body, &payload)
    if err != nil {
        return
    }

    err = json.Unmarshal(body, &inst)
    if err != nil {
        return
    }

    reg.Lock()
    defer reg.Unlock()

    key := fmt.Sprintf("%s.%s.%s", inst.Formation, inst.Service, inst.Instance)
    reg.entries[key] = Entry{payload, inst, time.Now()}
}

func (reg *Registry) HandleEvents() {
    input := make(chan map[string]interface{})
    go reg.serf.Stream("", input)

    for {
        data := <- input
        name, ok := data["Name"].(string)
        if ok && name == "advertise" {
            payload, _ := data["Payload"].([]byte)
            reg.addEntryFromEvent(payload)
        }
    }
}

func main() {
    serf, err := client.NewRPCClient("127.0.0.1:7373")
    if err != nil {
        log.Fatal(err)
    }
    reg := NewRegistry(serf)
    router := mux.NewRouter()
    router.HandleFunc("/{formation}", reg.QueryFormationHandler).Methods("GET")
    router.HandleFunc("/{formation}/{instance}", reg.UpdateHandler).Methods("PUT")

    go reg.HandleEvents()
	http.ListenAndServe(":4100", router)
}

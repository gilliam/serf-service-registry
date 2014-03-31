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

import "sync"
import "time"
import "errors"
//import "encoding/json"
//import "io/ioutil"

type JSONObject map[string]interface{}

type Entry struct {
    Payload   JSONObject
    Formation string
    Identity  string
    Timestamp time.Time
}

type IRegistry interface {
    Query(formation, identity string) (*Entry, error)
    Index(formation string) map[string]Entry
    Update(formation, identity string, payload JSONObject)
}

type Registry struct {
    sync.Mutex                  // lock for the registry
    entries    []Entry
}

func NewRegistry() *Registry {
    return &Registry{sync.Mutex{}, make([]Entry, 0, 1)}
}

func (registry *Registry) query(formation, identity string) (*Entry, error) {
    for _, entry := range registry.entries {
        if entry.Formation == formation && entry.Identity == identity {
            return &entry, nil
        }
    }
    return nil, errors.New("no such item")
}

func (registry *Registry) Query(formation, identity string) (*Entry, error) {
    registry.Lock()
    defer registry.Unlock()
    return registry.query(formation, identity)
}

func (registry *Registry) Index(formation string) map[string]Entry {
    registry.Lock()
    defer registry.Unlock()

    items := make(map[string]Entry)
    for _, entry := range registry.entries {
        if entry.Formation == formation {
            items[entry.Identity] = entry
        }
    }
    return items
}

func (registry *Registry) Update(formation, identity string, payload JSONObject) {
    registry.Lock()
    defer registry.Unlock()

    entry, err := registry.query(formation, identity)
    if err != nil {
        entry = &Entry{payload, formation, identity, time.Now()}
        registry.entries = append(registry.entries, *entry)
    }
    entry.Timestamp = time.Now()
}

// Author: lipixun
// Created Time : 日 12/11 19:09:51 2016
//
// File Name: trace.go
// Description:
//	The trace functions
package sourcecode

import (
	"fmt"
	"strings"
)

const (
	TraceTypeRepository = "Repository"
	TraceTypeTarget     = "Target"
	TraceTypeDependency = "Dependency"
)

type Tracer struct {
	path  []TraceItem
	types map[string]map[string]int
}

func NewTracer() *Tracer {
	return &Tracer{
		types: make(map[string]map[string]int),
	}
}

type TraceItem struct {
	Type string
	Key  string
	Name string
}

func (this *Tracer) Has(t, key string) bool {
	keys, ok := this.types[t]
	if !ok {
		return false
	}
	return keys[key] > 0
}

func (this *Tracer) Push(t, key, name string) {
	this.path = append(this.path, TraceItem{Type: t, Key: key, Name: name})
	keys, ok := this.types[t]
	if !ok {
		this.types[t] = map[string]int{key: 1}
	} else {
		keys[key] += 1
	}
}

func (this *Tracer) Pop() {
	if len(this.path) > 0 {
		item := this.path[len(this.path)-1]
		this.path = this.path[:len(this.path)-1]
		keys, ok := this.types[item.Type]
		if ok {
			value, ok := keys[item.Key]
			if ok {
				if value <= 1 {
					delete(keys, item.Key)
				} else {
					keys[item.Key] -= 1
				}
			}
		}
	}
}

func (this *Tracer) String() string {
	var strs []string
	for _, item := range this.path {
		strs = append(strs, fmt.Sprintf("%s:%s", item.Type, item.Name))
	}
	return strings.Join(strs, " --> ")
}

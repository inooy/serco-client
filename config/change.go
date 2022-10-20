package config

import (
	"github.com/inooy/serco-client/config/remote"
	"github.com/inooy/serco-client/pkg/log"
	"gopkg.in/yaml.v3"
	"reflect"
	"sync"
)

type PropChangeType int

const (
	ADDED    PropChangeType = 1
	MODIFIED PropChangeType = 2
	DELETED  PropChangeType = 3
)

type PropChangeEvent struct {
	ChangeType PropChangeType
	PropName   string
	OldValue   interface{}
	NewValue   interface{}
}

type PropChangeListener interface {
}

func calcChange(old *remote.Metadata, current *remote.Metadata) ([]*PropChangeEvent, error) {
	// 对比两个yaml/properties/json 格式的配置
	// 这里可以拿到原始的字符串
	// 解析为map，对比map差异？
	// 关于嵌套，这里不应该解析为嵌套
	var currentMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(current.Content), &currentMap); err != nil {
		log.Errorf("config handle error", err.Error())
		return nil, err
	}
	var oldMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(old.Content), &oldMap); err != nil {
		log.Errorf("config handle error", err.Error())
		return nil, err
	}
	flapCurrent := make(map[string]interface{})
	flapOld := make(map[string]interface{})
	flapMap(currentMap, flapCurrent, "")
	flapMap(oldMap, flapOld, "")
	return calcMap(flapOld, flapCurrent), nil
}

func flapMap(currentMap map[string]interface{}, result map[string]interface{}, keyPrefix string) {
	for key := range currentMap {
		subKey := key
		if keyPrefix != "" {
			subKey = keyPrefix + "." + key
		}
		switch reflect.TypeOf(currentMap[key]).Kind() {
		case reflect.Int:
			fallthrough
		case reflect.Int64:
			fallthrough
		case reflect.String:
			fallthrough
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			fallthrough
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			fallthrough
		case reflect.Bool:
			result[subKey] = currentMap[key]
		case reflect.Map:
			flapMap(currentMap[key].(map[string]interface{}), result, subKey)
		}
	}
}

func calcMap(oldMap map[string]interface{}, currentMap map[string]interface{}) []*PropChangeEvent {
	// 遍历对比map
	list := make([]*PropChangeEvent, 0)
	for key := range currentMap {
		if oldVal, ok := oldMap[key]; ok {
			//
			if !reflect.DeepEqual(currentMap[key], oldVal) {
				// 类型不同，那么需要触发对应key的事件
				event := PropChangeEvent{
					PropName:   key,
					ChangeType: MODIFIED,
					OldValue:   oldVal,
					NewValue:   currentMap[key],
				}
				list = append(list, &event)
			}
		} else {
			event := PropChangeEvent{
				PropName:   key,
				ChangeType: ADDED,
				OldValue:   nil,
				NewValue:   currentMap[key],
			}
			list = append(list, &event)
		}
	}
	for key := range oldMap {
		if _, ok := currentMap[key]; !ok {
			event := PropChangeEvent{
				PropName:   key,
				ChangeType: DELETED,
				OldValue:   oldMap[key],
				NewValue:   nil,
			}
			list = append(list, &event)
		}
	}
	return list
}

// PropEventEmitter k
// key: 可以是 a;可以是a.b; a.b.c
// 应该把  a ,a.b , a.b.c 都解析保存在set中，进行匹配触发事件
type PropEventEmitter struct {
	cLock     sync.RWMutex // protect the map
	callbacks map[string]func([]*PropChangeEvent)
}

func (emitter *PropEventEmitter) On(id string, callback func([]*PropChangeEvent)) {
	emitter.cLock.Lock()
	emitter.callbacks[id] = callback
	emitter.cLock.Unlock()
}

func (emitter *PropEventEmitter) Off(id string) {
	emitter.cLock.RLock()
	if _, ok := emitter.callbacks[id]; ok {
		delete(emitter.callbacks, id)
	}
	emitter.cLock.RUnlock()
}

func (emitter *PropEventEmitter) Emit(id string, dto []*PropChangeEvent) {
	if callback, ok := emitter.callbacks[id]; ok {
		callback(dto)
	}
}

package caspercloud

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ServerData struct {
	data   map[string][]Command
	index  map[string]Command
	lock   *sync.RWMutex
	random *rand.Rand
}

func NewServerData() *ServerData {
	return &ServerData{
		data:   make(map[string][]Command),
		index:  make(map[string]Command),
		lock:   &sync.RWMutex{},
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (self *ServerData) searchIdleCommand(cmds []Command) Command {
	for i := len(cmds) - 1; i > 0 && i > len(cmds)-5; i-- {
		log.Println(cmds[i].GetId(), cmds[i].Finished())
		if cmds[i].Finished() {
			continue
		}
		if cmds[i].GetStatus() == kCommandStatusIdle {
			log.Println("use idle command:", cmds[i].GetId())
			return cmds[i]
		}
	}
	return nil
}

func (self *ServerData) GetNewCommand(tmpl, proxyServer string) Command {
	log.Println(tmpl, self.data)
	self.lock.RLock()
	val, ok := self.data[tmpl]
	self.lock.RUnlock()
	if ok {
		log.Printf("find %d commands", len(val))
		c := self.searchIdleCommand(val)
		if c != nil {
			log.Println("return exist command", c.GetId())
			return c
		}
	} else {
		self.lock.Lock()
		var cmds []Command
		c := NewCasperCmd(tmpl+"_"+strconv.FormatInt(time.Now().UnixNano(), 10), tmpl, proxyServer)
		cmds = append(cmds, c)
		self.data[tmpl] = cmds
		self.index[c.GetId()] = c
		log.Println("add cmd for template:", tmpl)
		self.lock.Unlock()
		return c
	}

	c := NewCasperCmd(tmpl+"_"+strconv.FormatInt(time.Now().UnixNano(), 10), tmpl, proxyServer)
	val = append(val, c)
	self.index[c.GetId()] = c
	return c
}

func (self *ServerData) parseId(id string) (tmpl string, index int) {
	segs := strings.Split(id, "_")
	if len(segs) < 1 {
		return "", -1
	}
	tmpl = segs[0]
	index, _ = strconv.Atoi(segs[1])
	return tmpl, index
}

func (self *ServerData) GetCommand(id string) Command {
	self.lock.RLock()
	defer self.lock.RUnlock()

	log.Println("get cmd", self.index)
	val, ok := self.index[id]
	if !ok {
		return nil
	}
	return val
}

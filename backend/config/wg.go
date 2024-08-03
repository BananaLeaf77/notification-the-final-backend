package config

import "sync"

var waitGroupInstance *sync.WaitGroup

func GetWaitGroupInstance() *sync.WaitGroup {
	if waitGroupInstance == nil {
		waitGroupInstance = &sync.WaitGroup{}
	}
	return waitGroupInstance
}

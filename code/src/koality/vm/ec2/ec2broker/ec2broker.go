package ec2broker

import (
	"github.com/crowdmob/goamz/ec2"
	"sync"
	"time"
)

type EC2Broker struct {
	ec2             *ec2.EC2
	instancesCache  []ec2.Instance
	lastRequestTime time.Time
	cacheMutex      *sync.Mutex
	cacheExpiration time.Duration
}

func New(EC2 *ec2.EC2) *EC2Broker {
	var mutex sync.Mutex
	return &EC2Broker{
		ec2:             EC2,
		cacheMutex:      &mutex,
		cacheExpiration: time.Duration(5) * time.Second,
	}
}

// Not synchronized
func (broker *EC2Broker) updateInstanceCache() {
	instancesResp, err := broker.ec2.Instances(nil, nil)
	if err != nil {
		panic(err)
	}
	var instances []ec2.Instance
	for _, reservation := range instancesResp.Reservations {
		instances = append(instances, reservation.Instances...)
	}
	broker.instancesCache = instances
	broker.lastRequestTime = time.Now()
}

func (broker *EC2Broker) getLatestInstanceCache() []ec2.Instance {
	broker.cacheMutex.Lock()
	cacheIsStale := time.Now().After(broker.lastRequestTime.Add(broker.cacheExpiration))
	if cacheIsStale {
		broker.updateInstanceCache()
	}
	broker.cacheMutex.Unlock()
	return broker.instancesCache
}

func (broker *EC2Broker) Instances() []ec2.Instance {
	return broker.getLatestInstanceCache()
}

func (broker *EC2Broker) EC2() *ec2.EC2 {
	return broker.ec2
}

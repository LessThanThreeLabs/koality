package ec2broker

import (
	"github.com/LessThanThreeLabs/goamz/ec2"
	"sync"
	"time"
)

type Ec2Cache struct {
	EC2               *ec2.EC2
	reservationsCache []ec2.Reservation
	lastRequestTime   time.Time
	locker            sync.Locker
	expiration        time.Duration
}

// Not synchronized
func (cache *Ec2Cache) updateReservationsCache() {
	instancesResp, err := cache.EC2.Instances(nil, nil)
	if err != nil {
		panic(err)
	}
	cache.reservationsCache = instancesResp.Reservations
	cache.lastRequestTime = time.Now()
}

func (cache *Ec2Cache) getLatestReservationsCache() []ec2.Reservation {
	cache.locker.Lock()
	defer cache.locker.Unlock()

	cacheIsStale := time.Now().After(cache.lastRequestTime.Add(cache.expiration))
	if cacheIsStale {
		cache.updateReservationsCache()
	}
	return cache.reservationsCache
}

func (cache *Ec2Cache) Reservations() []ec2.Reservation {
	return cache.getLatestReservationsCache()
}

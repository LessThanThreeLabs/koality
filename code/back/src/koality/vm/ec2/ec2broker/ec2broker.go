package ec2broker

import (
	"github.com/crowdmob/goamz/ec2"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Ec2Broker struct {
	ec2               *ec2.EC2
	reservationsCache []ec2.Reservation
	lastRequestTime   time.Time
	cacheMutex        *sync.Mutex
	cacheExpiration   time.Duration
	instanceInfo      *InstanceInfo
}

type InstanceInfo struct {
	Name           string
	PrivateIp      net.IP
	SecurityGroups []ec2.SecurityGroup
}

func New(ec2 *ec2.EC2) *Ec2Broker {
	var mutex sync.Mutex
	return &Ec2Broker{
		ec2:             ec2,
		cacheMutex:      &mutex,
		cacheExpiration: time.Duration(5) * time.Second,
	}
}

// Not synchronized
func (broker *Ec2Broker) updateReservationsCache() {
	instancesResp, err := broker.ec2.Instances(nil, nil)
	if err != nil {
		panic(err)
	}
	broker.reservationsCache = instancesResp.Reservations
	broker.lastRequestTime = time.Now()
}

func (broker *Ec2Broker) getLatestReservationsCache() []ec2.Reservation {
	broker.cacheMutex.Lock()
	cacheIsStale := time.Now().After(broker.lastRequestTime.Add(broker.cacheExpiration))
	if cacheIsStale {
		broker.updateReservationsCache()
	}
	broker.cacheMutex.Unlock()
	return broker.reservationsCache
}

func (broker *Ec2Broker) Reservations() []ec2.Reservation {
	return broker.getLatestReservationsCache()
}

func (broker *Ec2Broker) Ec2() *ec2.EC2 {
	return broker.ec2
}

func (broker *Ec2Broker) InstanceInfo() *InstanceInfo {
	if broker.instanceInfo != nil {
		return broker.instanceInfo
	}
	reservations := broker.Reservations()

	var name string
	var securityGroups []ec2.SecurityGroup
	var privateIp net.IP

	instanceIdBytes, err := exec.Command("ec2metadata", "--instance-id").Output()
	if err == nil {
		instanceId := strings.TrimSpace(string(instanceIdBytes))
		for _, reservation := range reservations {
			for _, instance := range reservation.Instances {
				if instance.InstanceId == instanceId {
					for _, tag := range instance.Tags {
						if tag.Key == "Name" {
							name = tag.Value
						}
					}
					securityGroups = reservation.SecurityGroups
					privateIp = net.ParseIP(instance.PrivateIPAddress)
					break
				}
			}
		}
	}

	netInterfaces, err := net.Interfaces()
	if err != nil {
		panic(err) // This REALLY shouldn't happen
	}

	if name == "" {
		hostname, err := os.Hostname()
		if err != nil {
			name = "unknown"
		} else {
			name = hostname
		}
	}

	if privateIp == nil {
		for _, netInterface := range netInterfaces {
			if strings.HasPrefix(netInterface.Name, "eth0") || strings.HasPrefix(netInterface.Name, "en0") {
				addrs, err := netInterface.Addrs()
				if err != nil {
					panic(err) // This also REALLY shouldn't happen
				}

				for _, addr := range addrs {
					ip, _, err := net.ParseCIDR(addr.String())
					if err != nil {
						panic(err) // This also REALLY shouldn't happen
					}

					privateIp = ip.To4()
					if privateIp != nil {
						break
					}
				}
			}
		}
	}

	broker.instanceInfo = &InstanceInfo{name, privateIp, securityGroups}
	return broker.instanceInfo
}

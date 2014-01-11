package ec2broker

import (
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/ec2"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Ec2Broker struct {
	locker       sync.Locker
	caches       map[aws.Auth]Ec2Cache
	instanceInfo *InstanceInfo
}

type InstanceInfo struct {
	Name           string
	PrivateIp      net.IP
	SecurityGroups []ec2.SecurityGroup
}

func New() *Ec2Broker {
	return &Ec2Broker{
		locker: new(sync.Mutex),
		caches: make(map[aws.Auth]Ec2Cache),
	}
}

func (broker *Ec2Broker) Ec2Cache(auth aws.Auth) (*Ec2Cache, error) {
	broker.locker.Lock()
	defer broker.locker.Unlock()

	cache, ok := broker.caches[auth]
	if ok {
		return &cache, nil
	}

	region := aws.USWest2 // TODO (bbland): change to USEast
	availabilityZoneBytes, err := exec.Command("ec2metadata", "--availability-zone").Output()
	if err == nil {
		regionRegexp, err := regexp.Compile("(us|sa|eu|ap)-(north|south)?(east|west)?-[0-9]+")
		if err != nil {
			return nil, err // THIS SHOULDN'T HAPPEN
		}

		regionBytes := regionRegexp.Find(availabilityZoneBytes)
		foundRegion, ok := aws.Regions[string(regionBytes)]
		if ok {
			region = foundRegion
		}
	}
	ec2Conn := ec2.New(auth, region)

	// Validate the connection auth
	_, err = ec2Conn.Instances(nil, nil)
	if err != nil {
		return nil, err
	}

	cache = Ec2Cache{
		EC2:        ec2Conn,
		locker:     new(sync.Mutex),
		expiration: 5 * time.Second,
	}
	broker.caches[auth] = cache
	return &cache, nil
}

func (broker *Ec2Broker) InstanceInfo() *InstanceInfo {
	if broker.instanceInfo != nil {
		return broker.instanceInfo
	}
	if len(broker.caches) == 0 {
		panic("No aws credentials provided")
	}

	// Get an arbitrary ec2 cache
	var cache *Ec2Cache
	for _, ec2Cache := range broker.caches {
		cache = &ec2Cache
		break
	}

	reservations := cache.Reservations()

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

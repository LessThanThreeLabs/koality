package vm

import (
	"fmt"
	"sync"
)

// TODO: use debug-level logging everywhere

type virtualMachinePool struct {
	virtualMachineLauncher VirtualMachineLauncher
	minReady               uint64
	maxSize                uint64
	startingCount          uint64
	readyCount             int64 // This can be negative when overallocated
	allocatedCount         uint64
	readyChannel           chan VirtualMachine
	waitingChannel         chan chan VirtualMachine
	locker                 sync.Locker
}

func NewPool(virtualMachineLauncher VirtualMachineLauncher, minReady, maxSize uint64) *virtualMachinePool {
	if minReady > maxSize {
		panic(fmt.Sprintf("minReady should not be larger than maxSize: (%d > %d)", minReady, maxSize))
	}
	if maxSize == 0 {
		panic("maxSize must be positive: (was 0)")
	}
	pool := virtualMachinePool{
		virtualMachineLauncher: virtualMachineLauncher,
		minReady:               minReady,
		maxSize:                maxSize,
		readyChannel:           make(chan VirtualMachine, 64),
		waitingChannel:         make(chan chan VirtualMachine, 64),
		locker:                 new(sync.Mutex),
	}
	go pool.ensureReadyInstances()
	go pool.transferReadyToWaiting()
	return &pool
}

func (pool *virtualMachinePool) transferReadyToWaiting() {
	for waiting := range pool.waitingChannel {
		ready, ok := <-pool.readyChannel
		if !ok {
			panic("Ready channel closed")
		}
		waiting <- ready
	}
}

func (pool *virtualMachinePool) allocateN(numToAllocate uint64) {
	pool.locker.Lock()
	defer pool.locker.Unlock()

	pool.readyCount -= int64(numToAllocate)
	pool.allocatedCount += numToAllocate
	// fmt.Printf("Allocate %d: %#v\n", pool, numToAllocate)
}

func (pool *virtualMachinePool) unallocateOne() {
	pool.locker.Lock()
	defer pool.locker.Unlock()

	pool.allocatedCount--
}

func (pool *virtualMachinePool) ensureReadyInstances() error {
	numToLaunch := func() uint64 {
		pool.locker.Lock()
		defer pool.locker.Unlock()

		numToLaunch := int64(pool.minReady) - (pool.readyCount + int64(pool.startingCount))

		maxLaunchable := int64(pool.maxSize) - (pool.readyCount + int64(pool.startingCount+pool.allocatedCount))
		if numToLaunch > maxLaunchable {
			numToLaunch = maxLaunchable
		}
		// fmt.Printf("%#v\n", pool)
		// fmt.Printf("Num to launch: %d\n", numToLaunch)
		if numToLaunch <= 0 {
			return 0
		}
		pool.startingCount += uint64(numToLaunch)
		return uint64(numToLaunch)
	}()

	if numToLaunch == 0 {
		return nil
	}

	doneChannel := make(chan error, numToLaunch)

	for x := uint64(0); x < numToLaunch; x++ {
		go func(doneChannel chan error) {
			doneChannel <- pool.newReadyInstance()
		}(doneChannel)
	}

	for err := range doneChannel {
		if err != nil {
			return err
		}
	}

	return nil
}

func (pool *virtualMachinePool) Get() VirtualMachine {
	return <-pool.GetN(1)
}

func (pool *virtualMachinePool) GetN(numMachines uint64) <-chan VirtualMachine {
	machinesChan := make(chan VirtualMachine, numMachines)
	returnChan := make(chan VirtualMachine, numMachines)

	pool.allocateN(numMachines)

	// TODO (bbland): handle case where this errors out
	go pool.ensureReadyInstances()

	go func() {
		for x := uint64(0); x < numMachines; x++ {
			pool.waitingChannel <- machinesChan
		}
	}()

	// Necessary for closing the channel after numMachines are put on it
	go func() {
		for x := uint64(0); x < numMachines; x++ {
			returnChan <- <-machinesChan
		}
		close(returnChan)
	}()

	return returnChan
}

func (pool *virtualMachinePool) Free() {
	pool.unallocateOne()

	// TODO (bbland): handle case where this errors out
	go pool.ensureReadyInstances()
}

func (pool *virtualMachinePool) MaxSize() uint64 {
	return pool.maxSize
}

func (pool *virtualMachinePool) SetMaxSize(maxSize uint64) {
	pool.locker.Lock()
	defer pool.locker.Unlock()

	if maxSize < pool.minReady {
		panic(fmt.Sprintf("maxSize should not be smaller than minReady: (%d < %d)", maxSize, pool.minReady))
	}
	if maxSize == 0 {
		panic("maxSize must be positive: (was 0)")
	}

	pool.maxSize = maxSize

	numToRemove := int64(pool.startingCount+pool.allocatedCount) + pool.readyCount - int64(maxSize)
	if numToRemove <= 0 || pool.readyCount <= 0 {
		return
	}

	if numToRemove > pool.readyCount {
		numToRemove = pool.readyCount
	}

	for x := int64(0); x < numToRemove; x++ {
		select {
		case ready := <-pool.readyChannel:
			go ready.Terminate()
			pool.readyCount--
		default:
			panic("There wasn't actually a ready vm...")
		}
	}
}

func (pool *virtualMachinePool) MinReady() uint64 {
	return pool.minReady
}

func (pool *virtualMachinePool) SetMinReady(minReady uint64) {
	pool.locker.Lock()
	defer pool.locker.Unlock()

	if minReady > pool.maxSize {
		panic(fmt.Sprintf("minReady should not be larger than maxSize: (%d > %d)", minReady, pool.maxSize))
	}

	pool.minReady = minReady
	go pool.ensureReadyInstances()
}

func (pool *virtualMachinePool) newReadyInstance() error {
	newVm, err := func() (VirtualMachine, error) {
		// fmt.Println("try to launch")
		newVm, err := pool.virtualMachineLauncher.LaunchVirtualMachine()
		if err != nil {
			return nil, err
		}
		// fmt.Println("launched")

		pool.locker.Lock()
		defer pool.locker.Unlock()

		pool.startingCount--
		pool.readyCount++

		// fmt.Printf("New instance: %#v\n", pool)

		if pool.readyCount > int64(pool.minReady) {
			pool.readyCount--
			fmt.Println("Ermahgherd")
			return nil, newVm.Terminate()
		}
		return newVm, nil
	}()

	if err != nil {
		return err
	}
	if newVm == nil {
		// Probably shouldn't be panicing here, as this is a legitimate case when people shrink the pool
		panic("New Virtual Machine was nil")
	}

	pool.readyChannel <- newVm
	return nil
}

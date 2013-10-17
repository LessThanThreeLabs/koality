package vm

import (
	"fmt"
	"sync"
)

// TODO: use debug-level logging everywhere

type virtualMachinePool struct {
	virtualMachineLauncher VirtualMachineLauncher
	minReady  int
	maxSize   int
	startingCount int
	readyCount int
	allocatedCount int
	readyChannel chan VirtualMachine
	locker sync.Locker
}

func NewPool(virtualMachineLauncher VirtualMachineLauncher, minReady, maxSize int) *virtualMachinePool {
	readyChannel := make(chan VirtualMachine)
	locker := new(sync.Mutex)
	pool := virtualMachinePool{
		virtualMachineLauncher: virtualMachineLauncher,
		minReady: minReady,
		maxSize: maxSize,
		readyChannel: readyChannel,
		locker: locker,
	}
	go pool.ensureReadyInstances()
	return &pool
}

func (pool *virtualMachinePool) allocateOne() {
	pool.locker.Lock()
	defer pool.locker.Unlock()

	pool.readyCount--
	pool.allocatedCount++
	fmt.Printf("Allocate one: %#v\n", pool)
}

func (pool *virtualMachinePool) unallocateOne() {
	pool.locker.Lock()
	defer pool.locker.Unlock()

	pool.allocatedCount--
}

func (pool *virtualMachinePool) ensureReadyInstances() error {
	numToLaunch := func() int {
		pool.locker.Lock()
		defer pool.locker.Unlock()

		numToLaunch := pool.minReady - pool.readyCount - pool.startingCount
		fmt.Printf("%#v\n", pool)
		fmt.Printf("Num to launch: %d\n", numToLaunch)
		if numToLaunch <= 0 {
			return 0
		}
		pool.startingCount += numToLaunch
		return numToLaunch
	} ()

	if numToLaunch <= 0 {
		return nil
	}

	doneChannel := make(chan error)

	for x := 0; x < numToLaunch; x++ {
		go func(doneChannel chan error) {
			doneChannel <- pool.newReadyInstance()
		} (doneChannel)
	}

	for x := 0; x < numToLaunch; x++ {
		err := <-doneChannel
		if err != nil {
			return err
		}
	}

	return nil
}

func (pool *virtualMachinePool) Get() VirtualMachine {
	pool.allocateOne()

	go pool.ensureReadyInstances()

	return <-pool.readyChannel
}

func (pool *virtualMachinePool) Free() {
	pool.unallocateOne()

	pool.ensureReadyInstances()
}

func (pool *virtualMachinePool) MaxSize() int {
	return pool.maxSize
}

func (pool *virtualMachinePool) SetMaxSize(maxSize int) {
	pool.locker.Lock()
	defer pool.locker.Unlock()

	numToRemove := pool.startingCount + pool.readyCount + pool.allocatedCount - maxSize
	if numToRemove <= 0 || pool.readyCount <= 0 {
		return
	}

	if numToRemove > pool.readyCount {
		numToRemove = pool.readyCount
	}

	for x := 0; x < numToRemove; x++ {
		select {
		case ready := <-pool.readyChannel:
			go ready.Terminate()
			pool.readyCount--
		default:
			panic("There wasn't actually a ready vm...")
		}
	}
}

func (pool *virtualMachinePool) MinReady() int {
	return pool.minReady
}

func (pool *virtualMachinePool) SetMinReady(minReady int) {
	pool.locker.Lock()
	defer pool.locker.Unlock()

	pool.minReady = minReady
	go pool.ensureReadyInstances()
}

func (pool *virtualMachinePool) newReadyInstance() error {
	newVm, err := func() (VirtualMachine, error) {
		fmt.Println("try to launch")
		newVm := pool.virtualMachineLauncher.LaunchVirtualMachine()
		fmt.Println("launched")

		pool.locker.Lock()
		defer pool.locker.Unlock()

		pool.startingCount--
		pool.readyCount++

		fmt.Printf("New instance: %#v\n", pool)

		if pool.readyCount + pool.startingCount >= pool.maxSize {
			pool.readyCount--
			return nil, newVm.Terminate()
		}
		return newVm, nil
	}()

	if err != nil {
		return err
	}
	if newVm == nil {
		panic("New Virtual Machine was nil")
	}

	pool.readyChannel <- newVm
	return nil
}

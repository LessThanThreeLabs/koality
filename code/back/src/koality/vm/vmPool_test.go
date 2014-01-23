package vm_test

import (
	"koality/vm"
	"koality/vm/localmachine"
	"testing"
	"time"
)

const timeoutMultiplier = 500

func TestPoolSizeAssertions(test *testing.T) {
	poolSizeAllowed := func(minReady, maxSize uint64) (allowed bool) {
		defer func() {
			allowed = recover() == nil
		}()
		vm.NewPool(0, localmachine.Manager, minReady, maxSize)
		return
	}

	assertPoolSizeAllowed := func(minReady, maxSize uint64, shouldBeAllowed bool) {
		if poolSizeAllowed(minReady, maxSize) != shouldBeAllowed {
			if shouldBeAllowed {
				test.Errorf("Pool size not allowed, should be allowed, size params: (%d, %d)", minReady, maxSize)
			} else {
				test.Errorf("Pool size allowed, should not be allowed, size params: (%d, %d)", minReady, maxSize)
			}
		}
	}

	assertPoolSizeAllowed(0, 0, false)

	for maxSize := uint64(10); maxSize < 100; maxSize += 10 {
		assertPoolSizeAllowed(0, maxSize, true)
		assertPoolSizeAllowed(maxSize/2, maxSize, true)
		assertPoolSizeAllowed(maxSize, maxSize, true)
		assertPoolSizeAllowed(maxSize+1, maxSize, false)
		assertPoolSizeAllowed(-maxSize, maxSize, false)
	}

}

func TestPoolReachesCap(test *testing.T) {
	testPoolReachesCap := func(poolSize uint64) {
		pool := vm.NewPool(0, localmachine.Manager, 0, poolSize)
		timeout := time.After(time.Duration(poolSize*timeoutMultiplier) * time.Millisecond)

		vmChan, _ := pool.GetReady(poolSize)

		for x := uint64(0); x < poolSize; x++ {
			select {
			case <-timeout:
				test.Error("Timed out!")
			case machine := <-vmChan:
				if machine == nil {
					test.Error("Received nil from the pool")
				} else {
					machine.Terminate()
				}
			}
		}
	}

	testPoolReachesCap(1)
	testPoolReachesCap(5)
	testPoolReachesCap(77)
}

func TestPoolEnforcesCap(test *testing.T) {
	testPoolEnforcesCap := func(poolSize uint64) {
		pool := vm.NewPool(0, localmachine.Manager, 0, poolSize)
		timeout := time.After(time.Duration(poolSize*timeoutMultiplier) * time.Millisecond)

		vmChan, _ := pool.GetReady(poolSize + 1)

		for x := uint64(0); x < poolSize+1; x++ {
			select {
			case <-timeout:
				if x != poolSize {
					test.Error("Timed out unexpectedly!")
					// } else {
					// 	test.Log("Timed out as expected")
				}
			case machine, ok := <-vmChan:
				if machine != nil {
					defer machine.Terminate()
				}
				if x == poolSize && ok {
					test.Error("Got a machine when expected nil")
				} else if x != poolSize && !ok {
					test.Error("Got nil when expected a machine")
				}

			}
		}
	}

	testPoolEnforcesCap(1)
	testPoolEnforcesCap(5)
	testPoolEnforcesCap(77)
}

func TestPoolMaxSizeIncrease(test *testing.T) {
	testPoolMaxSizeIncrease := func(startingPoolSize, endingPoolSize uint64) {
		pool := vm.NewPool(0, localmachine.Manager, 0, startingPoolSize)
		timeout := time.After(time.Duration(startingPoolSize*timeoutMultiplier) * time.Millisecond)

		vmChan, _ := pool.GetReady(startingPoolSize)

		for x := uint64(0); x < startingPoolSize; x++ {
			select {
			case <-timeout:
				test.Error("Timed out!")
			case machine := <-vmChan:
				if machine == nil {
					test.Error("Received nil from the pool")
				} else {
					machine.Terminate()
				}
			}
		}

		pool.SetMaxSize(endingPoolSize)

		timeout = time.After(time.Duration((endingPoolSize-startingPoolSize)*timeoutMultiplier) * time.Millisecond)
		vmChan, _ = pool.GetReady(endingPoolSize - startingPoolSize)

		for x := startingPoolSize; x < endingPoolSize+1; x++ {
			select {
			case <-timeout:
				if x != endingPoolSize {
					test.Error("Timed out unexpectedly!")
				}
			case machine, ok := <-vmChan:
				if machine != nil {
					defer machine.Terminate()
				}
				if x == endingPoolSize && ok {
					test.Error("Got a machine when expected nil")
				} else if x != endingPoolSize && !ok {
					test.Error("Got nil when expected a machine")
				}
			}
		}
	}

	testPoolMaxSizeIncrease(1, 2)
	testPoolMaxSizeIncrease(1, 10)

	testPoolMaxSizeIncrease(7, 8)
	testPoolMaxSizeIncrease(7, 10)

	testPoolMaxSizeIncrease(77, 100)
}

func TestPoolMaxSizeDecrease(test *testing.T) {
	testPoolMaxSizeDecrease := func(startingPoolSize, endingPoolSize, amountToRequest uint64) {
		pool := vm.NewPool(0, localmachine.Manager, 0, startingPoolSize)
		timeout := time.After(time.Duration(startingPoolSize*timeoutMultiplier) * time.Millisecond)

		vmChan, _ := pool.GetReady(amountToRequest)

		for x := uint64(0); x < amountToRequest; x++ {
			select {
			case <-timeout:
				test.Error("Timed out!")
			case machine := <-vmChan:
				if machine == nil {
					test.Error("Received nil from the pool")
				} else {
					machine.Terminate()
				}
			}
		}

		pool.SetMaxSize(endingPoolSize)

		if amountToRequest <= endingPoolSize {
			timeout = time.After(time.Duration((endingPoolSize-amountToRequest)*timeoutMultiplier) * time.Millisecond)
			vmChan, _ = pool.GetReady(endingPoolSize - amountToRequest)

			for x := amountToRequest; x < endingPoolSize+1; x++ {
				select {
				case <-timeout:
					if x != endingPoolSize {
						test.Error("Timed out unexpectedly!")
					}
				case machine, ok := <-vmChan:
					if machine != nil {
						defer machine.Terminate()
					}
					if x == endingPoolSize && ok {
						test.Error("Got a machine when expected nil")
					} else if x != startingPoolSize && !ok {
						test.Error("Got nil when expected a machine")
					}
				}
			}
		} else {
			timeout = time.After(20 * time.Millisecond)
			vmChan, _ = pool.GetReady(1)

			select {
			case <-timeout:
				break
			case <-vmChan:
				test.Error("Got a machine when expected nil")
			}
		}
	}

	testPoolMaxSizeDecrease(3, 1, 2)
}

func TestPoolMinReadyIncrease(test *testing.T) {
	test.Skip("This is very important to add")
}

func TestPoolMinReadyDecrease(test *testing.T) {
	test.Skip("This is very important to add")
}

package localmachine

import (
	// "fmt"
	// "koality/shell"
	"koality/vm"
	"testing"
	"time"
)

func TestPoolSizeAssertions(testing *testing.T) {
	poolSizeAllowed := func(minReady, maxSize uint64) (allowed bool) {
		defer func() {
			allowed = recover() == nil
		}()
		vm.NewPool(NewLauncher(), minReady, maxSize)
		return
	}

	assertPoolSizeAllowed := func(minReady, maxSize uint64, shouldBeAllowed bool) {
		if poolSizeAllowed(minReady, maxSize) != shouldBeAllowed {
			if shouldBeAllowed {
				testing.Errorf("Pool size not allowed, should be allowed, size params: (%d, %d)", minReady, maxSize)
			} else {
				testing.Errorf("Pool size allowed, should not be allowed, size params: (%d, %d)", minReady, maxSize)
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

func TestPoolReachesCap(testing *testing.T) {
	testPoolReachesCap := func(poolSize uint64) {
		pool := vm.NewPool(NewLauncher(), 0, poolSize)
		timeout := time.After(time.Duration(poolSize*20) * time.Millisecond)

		vmChan := pool.GetN(poolSize)

		for x := uint64(0); x < poolSize; x++ {
			select {
			case <-timeout:
				testing.Error("Timed out!")
			case machine := <-vmChan:
				if machine == nil {
					testing.Error("Received nil from the pool")
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

func TestPoolEnforcesCap(testing *testing.T) {
	testPoolEnforcesCap := func(poolSize uint64) {
		pool := vm.NewPool(NewLauncher(), 0, poolSize)
		timeout := time.After(time.Duration(poolSize*20) * time.Millisecond)

		vmChan := pool.GetN(poolSize + 1)

		for x := uint64(0); x < poolSize+1; x++ {
			select {
			case <-timeout:
				if x != poolSize {
					testing.Error("Timed out unexpectedly!")
					// } else {
					// 	testing.Log("Timed out as expected")
				}
			case machine, ok := <-vmChan:
				if machine != nil {
					defer machine.Terminate()
				}
				if x == poolSize && ok {
					testing.Error("Got a machine when expected nil")
				} else if x != poolSize && !ok {
					testing.Error("Got nil when expected a machine")
				}

			}
		}
	}

	testPoolEnforcesCap(1)
	testPoolEnforcesCap(5)
	testPoolEnforcesCap(77)
}

func TestPoolMaxSizeIncrease(testing *testing.T) {
	testPoolMaxSizeIncrease := func(startingPoolSize, endingPoolSize uint64) {
		pool := vm.NewPool(NewLauncher(), 0, startingPoolSize)
		timeout := time.After(time.Duration(startingPoolSize*20) * time.Millisecond)

		vmChan := pool.GetN(startingPoolSize)

		for x := uint64(0); x < startingPoolSize; x++ {
			select {
			case <-timeout:
				testing.Error("Timed out!")
			case machine := <-vmChan:
				if machine == nil {
					testing.Error("Received nil from the pool")
				} else {
					machine.Terminate()
				}
			}
		}

		pool.SetMaxSize(endingPoolSize)

		timeout = time.After(time.Duration((endingPoolSize-startingPoolSize)*20) * time.Millisecond)
		vmChan = pool.GetN(endingPoolSize - startingPoolSize)

		for x := startingPoolSize; x < endingPoolSize+1; x++ {
			select {
			case <-timeout:
				if x != endingPoolSize {
					testing.Error("Timed out unexpectedly!")
				}
			case machine, ok := <-vmChan:
				if machine != nil {
					defer machine.Terminate()
				}
				if x == endingPoolSize && ok {
					testing.Error("Got a machine when expected nil")
				} else if x != endingPoolSize && !ok {
					testing.Error("Got nil when expected a machine")
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

func TestPoolMaxSizeDecrease(testing *testing.T) {
	testPoolMaxSizeDecrease := func(startingPoolSize, endingPoolSize, amountToRequest uint64) {
		pool := vm.NewPool(NewLauncher(), 0, startingPoolSize)
		timeout := time.After(time.Duration(startingPoolSize*20) * time.Millisecond)

		vmChan := pool.GetN(amountToRequest)

		for x := uint64(0); x < amountToRequest; x++ {
			select {
			case <-timeout:
				testing.Error("Timed out!")
			case machine := <-vmChan:
				if machine == nil {
					testing.Error("Received nil from the pool")
				} else {
					machine.Terminate()
				}
			}
		}

		pool.SetMaxSize(endingPoolSize)

		if amountToRequest <= endingPoolSize {
			timeout = time.After(time.Duration((endingPoolSize-amountToRequest)*20) * time.Millisecond)
			vmChan = pool.GetN(endingPoolSize - amountToRequest)

			for x := amountToRequest; x < endingPoolSize+1; x++ {
				select {
				case <-timeout:
					if x != endingPoolSize {
						testing.Error("Timed out unexpectedly!")
					}
				case machine, ok := <-vmChan:
					if machine != nil {
						defer machine.Terminate()
					}
					if x == endingPoolSize && ok {
						testing.Error("Got a machine when expected nil")
					} else if x != startingPoolSize && !ok {
						testing.Error("Got nil when expected a machine")
					}
				}
			}
		} else {
			timeout = time.After(20 * time.Millisecond)
			vmChan = pool.GetN(1)

			select {
			case <-timeout:
				break
			case <-vmChan:
				testing.Error("Got a machine when expected nil")
			}
		}
	}

	testPoolMaxSizeDecrease(3, 1, 2)
}

func TestPoolMinReadyIncrease(testing *testing.T) {
	testing.Log("This is very important to add")
}

func TestPoolMinReadyDecrease(testing *testing.T) {
	testing.Log("This is very important to add")
}

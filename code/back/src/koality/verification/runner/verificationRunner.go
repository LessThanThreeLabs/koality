package runner

import (
	"fmt"
	"koality/repositorymanager"
	"koality/repositorymanager/pathgenerator"
	"koality/resources"
	"koality/shell"
	"koality/verification"
	"koality/verification/config"
	"koality/verification/config/commandgroup"
	"koality/verification/config/section"
	"koality/verification/stagerunner"
	"koality/vm"
	"koality/vm/ec2/ec2broker"
	"koality/vm/ec2/ec2vm"
	"koality/vm/vcs"
	"sync"
	"time"
)

type VerificationRunner struct {
	resourcesConnection                  *resources.Connection
	virtualMachinePoolMap                map[uint64]vm.VirtualMachinePool
	virtualMachinePoolMapLocker          sync.Locker
	ec2Broker                            *ec2broker.Ec2Broker
	verificationCreatedSubscriptionId    resources.SubscriptionId
	ec2PoolCreatedSubscriptionId         resources.SubscriptionId
	ec2PoolDeletedSubscriptionId         resources.SubscriptionId
	ec2PoolSettingsUpdatedSubscriptionId resources.SubscriptionId
}

func New(resourcesConnection *resources.Connection, virtualMachinePools []vm.VirtualMachinePool, ec2Broker *ec2broker.Ec2Broker) *VerificationRunner {
	virtualMachinePoolMap := make(map[uint64]vm.VirtualMachinePool, len(virtualMachinePools))
	for _, virtualMachinePool := range virtualMachinePools {
		virtualMachinePoolMap[virtualMachinePool.Id()] = virtualMachinePool
	}

	return &VerificationRunner{
		resourcesConnection:         resourcesConnection,
		virtualMachinePoolMap:       virtualMachinePoolMap,
		virtualMachinePoolMapLocker: new(sync.Mutex),
		ec2Broker:                   ec2Broker,
	}
}

func (verificationRunner *VerificationRunner) SubscribeToEvents() error {
	onVerificationCreated := func(verification *resources.Verification) {
		verificationRunner.RunVerification(verification)
	}
	onEc2PoolCreated := func(ec2Pool *resources.Ec2Pool) {
		verificationRunner.virtualMachinePoolMapLocker.Lock()
		ec2VirtualMachinePool := ec2vm.NewPool(ec2vm.NewLauncher(verificationRunner.ec2Broker, ec2Pool))
		verificationRunner.virtualMachinePoolMap[ec2Pool.Id] = ec2VirtualMachinePool
		verificationRunner.virtualMachinePoolMapLocker.Unlock()
	}
	onEc2PoolDeleted := func(ec2PoolId uint64) {
		verificationRunner.virtualMachinePoolMapLocker.Lock()
		delete(verificationRunner.virtualMachinePoolMap, ec2PoolId)
		verificationRunner.virtualMachinePoolMapLocker.Unlock()
	}
	onEc2PoolSettingsUpdated := func(ec2PoolId uint64, accessKey, secretKey, username, baseAmiId, securityGroupId,
		vpcSubnetId, instanceType string, numReadyInstances, numMaxInstances, rootDriveSize uint64, userData string) {
		verificationRunner.virtualMachinePoolMapLocker.Lock()
		vmPool, ok := verificationRunner.virtualMachinePoolMap[ec2PoolId]
		verificationRunner.virtualMachinePoolMapLocker.Unlock()
		if !ok {
			panic(fmt.Sprintf("Could not find pool with id: %d", ec2PoolId))
		}

		ec2VmPool, ok := vmPool.(ec2vm.Ec2VirtualMachinePool)
		if !ok {
			panic(fmt.Sprintf("Pool with id: %d is not an EC2 pool", ec2PoolId))
		}

		ec2Pool := ec2VmPool.Ec2VirtualMachineLauncher.Ec2Pool

		ec2VmPool.UpdateSettings(resources.Ec2Pool{ec2PoolId, ec2Pool.Name, accessKey, secretKey, username, baseAmiId,
			securityGroupId, vpcSubnetId, instanceType, numReadyInstances, numMaxInstances, rootDriveSize, userData, ec2Pool.Created})
	}
	var err error

	verificationRunner.verificationCreatedSubscriptionId, err = verificationRunner.resourcesConnection.Verifications.Subscription.SubscribeToCreatedEvents(onVerificationCreated)
	if err != nil {
		return err
	}

	verificationRunner.ec2PoolCreatedSubscriptionId, err = verificationRunner.resourcesConnection.Pools.Subscription.SubscribeToEc2CreatedEvents(onEc2PoolCreated)
	if err != nil {
		verificationRunner.unsubscribeFromEvents(true)
		return err
	}

	verificationRunner.ec2PoolDeletedSubscriptionId, err = verificationRunner.resourcesConnection.Pools.Subscription.SubscribeToEc2DeletedEvents(onEc2PoolDeleted)
	if err != nil {
		verificationRunner.unsubscribeFromEvents(true)
		return err
	}

	verificationRunner.ec2PoolSettingsUpdatedSubscriptionId, err = verificationRunner.resourcesConnection.Pools.Subscription.SubscribeToEc2SettingsUpdatedEvents(onEc2PoolSettingsUpdated)
	if err != nil {
		verificationRunner.unsubscribeFromEvents(true)
		return err
	}
	return nil
}

func (verificationRunner *VerificationRunner) UnsubscribeFromEvents() error {
	return verificationRunner.unsubscribeFromEvents(false)
}

func (verificationRunner *VerificationRunner) unsubscribeFromEvents(allowPartial bool) error {
	var err error

	if verificationRunner.verificationCreatedSubscriptionId == 0 {
		if !allowPartial {
			return fmt.Errorf("Verification created events not subscribed to")
		}
	} else {
		unsubscribeError := verificationRunner.resourcesConnection.Verifications.Subscription.UnsubscribeFromCreatedEvents(verificationRunner.verificationCreatedSubscriptionId)
		if unsubscribeError != nil {
			err = unsubscribeError
		}
	}
	if verificationRunner.ec2PoolCreatedSubscriptionId == 0 {
		if !allowPartial {
			return fmt.Errorf("Ec2 pool created events not subscribed to")
		}
	} else {
		unsubscribeError := verificationRunner.resourcesConnection.Pools.Subscription.UnsubscribeFromEc2CreatedEvents(verificationRunner.ec2PoolCreatedSubscriptionId)
		if unsubscribeError != nil {
			err = unsubscribeError
		}
	}
	if verificationRunner.ec2PoolDeletedSubscriptionId == 0 {
		if !allowPartial {
			return fmt.Errorf("Ec2 pool deleted events not subscribed to")
		}
	} else {
		unsubscribeError := verificationRunner.resourcesConnection.Pools.Subscription.UnsubscribeFromEc2DeletedEvents(verificationRunner.ec2PoolDeletedSubscriptionId)
		if unsubscribeError != nil {
			err = unsubscribeError
		}
	}
	if verificationRunner.ec2PoolSettingsUpdatedSubscriptionId == 0 {
		if !allowPartial {
			return fmt.Errorf("Ec2 pool settings updated events not subscribed to")
		}
	} else {
		unsubscribeError := verificationRunner.resourcesConnection.Pools.Subscription.UnsubscribeFromEc2SettingsUpdatedEvents(verificationRunner.ec2PoolSettingsUpdatedSubscriptionId)
		if unsubscribeError != nil {
			err = unsubscribeError
		}
	}
	return err
}

func (verificationRunner *VerificationRunner) RunVerification(currentVerification *resources.Verification) (bool, error) {
	if currentVerification == nil {
		panic("Cannot run a nil verification")
	}
	verificationConfig, err := verificationRunner.getVerificationConfig(currentVerification)
	if err != nil {
		verificationRunner.failVerification(currentVerification)
		return false, err
	}

	verificationRunner.virtualMachinePoolMapLocker.Lock()
	virtualMachinePool, ok := verificationRunner.virtualMachinePoolMap[verificationConfig.Params.PoolId]
	verificationRunner.virtualMachinePoolMapLocker.Unlock()
	if !ok {
		verificationRunner.failVerification(currentVerification)
		return false, fmt.Errorf("No virtual machine pool found for repository: %d", currentVerification.RepositoryId)
	}

	err = verificationRunner.createStages(currentVerification, verificationConfig.Sections, verificationConfig.FinalSections)
	if err != nil {
		verificationRunner.failVerification(currentVerification)
		return false, err
	}

	numNodes := verificationConfig.Params.Nodes
	if numNodes == 0 {
		numNodes = 1 // TODO (bbland): Do a better job guessing the number of nodes when unspecified
	}

	newStageRunnersChan := make(chan *stagerunner.StageRunner, numNodes)

	runStages := func(virtualMachine vm.VirtualMachine) {
		defer virtualMachinePool.Free()
		defer virtualMachine.Terminate()

		stageRunner := stagerunner.New(verificationRunner.resourcesConnection, virtualMachine, currentVerification)
		defer close(stageRunner.ResultsChan)

		newStageRunnersChan <- stageRunner

		err := stageRunner.RunStages(verificationConfig.Sections, verificationConfig.FinalSections, verificationConfig.Params.Environment)
		if err != nil {
			panic(err)
		}
	}

	newMachinesChan, errorChan := virtualMachinePool.Get(numNodes)
	go func(newMachinesChan <-chan vm.VirtualMachine) {
		for newMachine := range newMachinesChan {
			go runStages(newMachine)
		}
	}(newMachinesChan)

	// TODO (bbland): do something with the errorChan
	go func(errorChan <-chan error) {
		<-errorChan
	}(errorChan)

	resultsChan := verificationRunner.combineResults(currentVerification, newStageRunnersChan)
	receivedResult := make(map[string]bool)

	for {
		result, hasMoreResults := <-resultsChan
		if !hasMoreResults {
			verificationPassed := currentVerification.Status != "cancelled" && currentVerification.Status != "failed"
			if verificationPassed {
				err := verificationRunner.passVerification(currentVerification)
				if err != nil {
					return false, err
				}
			}
			return verificationPassed, nil
		}
		if result.Passed == false {
			switch result.FailSectionOn {
			case section.FailOnNever:
				// Do nothing
			case section.FailOnAny:
				if currentVerification.Status != "cancelled" && currentVerification.Status != "failed" {
					err := verificationRunner.failVerification(currentVerification)
					if err != nil {
						return false, err
					}
				}
			case section.FailOnFirst:
				if !receivedResult[result.Section] {
					if currentVerification.Status != "cancelled" && currentVerification.Status != "failed" {
						err := verificationRunner.failVerification(currentVerification)
						if err != nil {
							return false, err
						}
					}
				}
			}
		}
		receivedResult[result.Section] = true
	}
}

func (verificationRunner *VerificationRunner) failVerification(verification *resources.Verification) error {
	(*verification).Status = "failed"
	fmt.Printf("verification %d %sFAILED!!!%s\n", verification.Id, shell.AnsiFormat(shell.AnsiFgRed, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	err := verificationRunner.resourcesConnection.Verifications.Update.SetStatus(verification.Id, "failed")
	if err != nil {
		return err
	}

	err = verificationRunner.resourcesConnection.Verifications.Update.SetEndTime(verification.Id, time.Now())
	return err
}

func (verificationRunner *VerificationRunner) passVerification(verification *resources.Verification) error {
	(*verification).Status = "passed"
	fmt.Printf("verification %d %sPASSED!!!%s\n", verification.Id, shell.AnsiFormat(shell.AnsiFgGreen, shell.AnsiBold), shell.AnsiFormat(shell.AnsiReset))
	err := verificationRunner.resourcesConnection.Verifications.Update.SetStatus(verification.Id, "passed")
	if err != nil {
		return err
	}

	err = verificationRunner.resourcesConnection.Verifications.Update.SetEndTime(verification.Id, time.Now())
	return err
}

func (verificationRunner *VerificationRunner) getVerificationConfig(currentVerification *resources.Verification) (config.VerificationConfig, error) {
	var emptyConfig config.VerificationConfig

	repository, err := verificationRunner.resourcesConnection.Repositories.Read.Get(currentVerification.RepositoryId)
	if err != nil {
		return emptyConfig, err
	}

	configYaml, err := repositorymanager.GetYamlFile(repository, currentVerification.Changeset.HeadSha)
	if err != nil {
		panic(err)
		return emptyConfig, err
	}

	verificationConfig, err := config.FromYaml(configYaml)
	if err != nil {
		return emptyConfig, err
	}

	// TODO (bbland): add retry logic
	checkoutCommand := vcs.CheckoutCommand(repository, pathgenerator.GitHiddenRef(currentVerification.Changeset.HeadSha))
	setupCommands := []verification.Command{verification.NewShellCommand(repository.VcsType, checkoutCommand)}
	setupSection := section.New("setup", section.RunOnAll, section.FailOnFirst, false, nil, commandgroup.New(setupCommands), nil)
	verificationConfig.Sections = append([]section.Section{setupSection}, verificationConfig.Sections...)
	return verificationConfig, nil
}

func (verificationRunner *VerificationRunner) createStages(currentVerification *resources.Verification, sections, finalSections []section.Section) error {
	for sectionNumber, section := range append(sections, finalSections...) {
		var err error

		stageNumber := 0

		factoryCommands := section.FactoryCommands(true)

		for command, err := factoryCommands.Next(); err == nil; command, err = factoryCommands.Next() {
			_, err := verificationRunner.resourcesConnection.Stages.Create.Create(currentVerification.Id, uint64(sectionNumber), command.Name(), uint64(stageNumber))
			if err != nil {
				return err
			}
			stageNumber++
		}
		if err != nil && err != commandgroup.NoMoreCommands {
			return err
		}

		commands := section.Commands(true)
		for command, err := commands.Next(); err == nil; command, err = commands.Next() {
			_, err := verificationRunner.resourcesConnection.Stages.Create.Create(currentVerification.Id, uint64(sectionNumber), command.Name(), uint64(stageNumber))
			if err != nil {
				return err
			}
			stageNumber++
		}
		if err != nil && err != commandgroup.NoMoreCommands {
			return err
		}
	}
	return nil
}

func (verificationRunner *VerificationRunner) combineResults(currentVerification *resources.Verification, newStageRunnersChan <-chan *stagerunner.StageRunner) <-chan verification.SectionResult {
	resultsChan := make(chan verification.SectionResult)
	go func(newStageRunnersChan <-chan *stagerunner.StageRunner) {
		isStarted := false

		combinedResults := make(chan verification.SectionResult)
		stageRunners := make([]*stagerunner.StageRunner, 0, cap(newStageRunnersChan))
		stageRunnerDoneChan := make(chan error, cap(newStageRunnersChan))
		stageRunnersDoneCounter := 0

		handleNewStageRunner := func(stageRunner *stagerunner.StageRunner) {
			for {
				result, ok := <-stageRunner.ResultsChan
				if !ok {
					stageRunnerDoneChan <- nil
					return
				}
				combinedResults <- result
			}
		}

		// TODO (bbland): make this not so ugly
	gatherResults:
		for {
			select {
			case stageRunner, ok := <-newStageRunnersChan:
				if !ok {
					panic("new stage Runners channel closed")
				}
				if !isStarted {
					isStarted = true
					verificationRunner.resourcesConnection.Verifications.Update.SetStartTime(currentVerification.Id, time.Now())
				}
				stageRunners = append(stageRunners, stageRunner)
				go handleNewStageRunner(stageRunner)

			case err, ok := <-stageRunnerDoneChan:
				if !ok {
					panic("stage runner done channel closed")
				}
				if err != nil {
					panic(err)
				}
				stageRunnersDoneCounter++
				if stageRunnersDoneCounter >= len(stageRunners) {
					break gatherResults
				}

			case result, ok := <-combinedResults:
				if !ok {
					panic("combined results channel closed")
				}
				resultsChan <- result
			}
		}
		// Drain the remaining results that matter
		drainRemaining := func() {
			for {
				select {
				case result := <-combinedResults:
					resultsChan <- result
				default:
					close(resultsChan)
					return
				}
			}
		}
		drainRemaining()
		// Drain any extra possible results that we don't care about
		for _, stageRunner := range stageRunners {
			for _ = range stageRunner.ResultsChan {
				fmt.Println("Got an extra result")
			}
		}
	}(newStageRunnersChan)
	return resultsChan
}

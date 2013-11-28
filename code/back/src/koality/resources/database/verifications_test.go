package database

import (
	"fmt"
	"testing"
)

const (
	verificationRepositoryName      = "repository-name"
	verificationRepositoryVcsType   = "git"
	verificationRepositoryLocalUri  = "git@local.uri.com:name.git"
	verificationRepositoryRemoteUri = "git@remote.uri.com:name.git"
	headSha                         = "a5a1134e5ca1050a2ea01b1b8a9f945bc758ec49"
	baseSha                         = "5984b365f6a7287d8b3673b200525bb769a5a3de"
	headMessage                     = "This is an awesome commit message"
	headUsername                    = "Jordan Potter"
	headEmail                       = "jpotter@koalitycode.com"
	mergeTarget                     = "refs/heads/master"
	emailToNotify                   = "koalas@koalitycode.com"
)

func TestCreateInvalidVerification(test *testing.T) {

}

func TestCreateVerification(test *testing.T) {
	connection, err := New()
	if err != nil {
		test.Fatal(err)
	}

	repositoryId, err := connection.Repositories.Create.Create(verificationRepositoryName, verificationRepositoryVcsType, verificationRepositoryLocalUri, verificationRepositoryRemoteUri)
	if err != nil {
		test.Fatal(err)
	}

	verificationId, err := connection.Verifications.Create.Create(repositoryId, headSha, baseSha, headMessage, headUsername, headEmail, mergeTarget, emailToNotify)
	if err != nil {
		test.Fatal(err)
	}

	fmt.Println(verificationId)

	// repository, err := connection.Repositories.Read.Get(repositoryId)
	// if err != nil {
	// 	test.Fatal(err)
	// }

	// if repository.Id != repositoryId {
	// 	test.Fatal("repository.Id mismatch")
	// }

	// _, err = connection.Repositories.Create.Create(repository.Name, repository.VcsType, repository.LocalUri, repository.RemoteUri)
	// if _, ok := err.(resources.RepositoryAlreadyExistsError); !ok {
	// 	test.Fatal("Expected RepositoryAlreadyExistsError when trying to add same repository twice")
	// }

	// err = connection.Repositories.Update.SetGitHubHook(repositoryId, gitHubHookId, gitHubHookSecret, gitHubHookTypes)
	// if _, ok := err.(resources.NoSuchRepositoryHookError); !ok {
	// 	test.Fatal("Expected NoSuchRepositoryHookError when trying to add repository hook")
	// }

	// err = connection.Repositories.Update.ClearGitHubHook(repositoryId)
	// if _, ok := err.(resources.NoSuchRepositoryHookError); !ok {
	// 	test.Fatal("Expected NoSuchRepositoryHookError when trying to clear repository hook")
	// }

	// err = connection.Repositories.Delete.Delete(repositoryId)
	// if err != nil {
	// 	test.Fatal(err)
	// }

	// err = connection.Repositories.Delete.Delete(repositoryId)
	// if _, ok := err.(resources.NoSuchRepositoryError); !ok {
	// 	test.Fatal("Expected NoSuchRepositoryError when trying to delete same repository twice")
	// }
}

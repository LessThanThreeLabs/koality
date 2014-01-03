package pathgenerator

import (
	"fmt"
	"koality/resources"
	"path/filepath"
	"strconv"
)

func ToPath(repo *resources.Repository) string {
	return filepath.Join("/", "etc", "repositories", fmt.Sprintf("%d", repo.Id), repo.Name)
}

func GitHiddenRef(commitId string) string {
	return fmt.Sprintf("refs/pending/%s", commitId)
}

func GetRepoID(absPath string) (int64, error) {
	list := filepath.SplitList(absPath)

	repoId, err := strconv.ParseInt(list[len(list)-2], 10, 64)

	if err != nil {
		return 0, fmt.Errorf("Invalid repo path %s. This repository was not stored by the repostore.", absPath)
	}

	return repoId, nil
}

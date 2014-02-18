package licenseserver

import (
	"encoding/json"
	"fmt"
	"koality/license"
	"net/http"
	"regexp"
)

var upgradeMatchRegex = regexp.MustCompile("/koality-((?:\\d+\\.)*\\d+)\\.tgz$")

func (licenseServer *LicenseServer) checkUpgrade(writer http.ResponseWriter, request *http.Request) {
	checkUpgradeRequest := new(license.CheckUpgradeRequest)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(checkUpgradeRequest); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if checkUpgradeRequest.LicenseKey == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a license key")
		return
	} else if checkUpgradeRequest.ServerId == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a server id")
		return
	} else if checkUpgradeRequest.CurrentVersion == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a current version")
		return
	}

	licenseCheckRequest := &license.CheckLicenseRequest{checkUpgradeRequest.LicenseKey, checkUpgradeRequest.ServerId}

	_, err := licenseServer.checkLicenseKey(licenseCheckRequest)
	if err == license.NotFoundError || err == license.DeactivatedError || err == license.ServerIdMismatchError {
		licenseServer.printCheckUpgradeResponse(new(license.CheckUpgradeResponse), writer, request)
		return
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	checkUpgradeResponse, err := licenseServer.getUpgradeInfo(checkUpgradeRequest.CurrentVersion)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	licenseServer.printCheckUpgradeResponse(checkUpgradeResponse, writer, request)
}

func (licenseServer *LicenseServer) getUpgradeInfo(currentVersion string) (*license.CheckUpgradeResponse, error) {
	changelog, err := licenseServer.getChangelog()
	if err != nil {
		return nil, err
	}

	latestVersion, err := licenseServer.getLatestVersion()
	if err != nil {
		return nil, err
	}

	if latestVersion <= currentVersion {
		return &license.CheckUpgradeResponse{
			HasUpgrade: false,
		}, nil
	}

	filteredChangelog := make([]license.ChangeInfo, 0, len(changelog))
	for _, changeInfo := range changelog {
		if currentVersion < changeInfo.VersionAdded && changeInfo.VersionAdded <= latestVersion {
			filteredChangelog = append(filteredChangelog, changeInfo)
		}
	}

	return &license.CheckUpgradeResponse{true, latestVersion, filteredChangelog}, nil
}

func (licenseServer *LicenseServer) getChangelog() (license.Changelog, error) {
	changelogData, err := licenseServer.bucket.Get(fmt.Sprintf("%s/changelog.json", upgradePathPrefix))
	if err != nil {
		return nil, err
	}

	var changelog license.Changelog
	if err = json.Unmarshal(changelogData, &changelog); err != nil {
		return nil, err
	}

	return changelog, nil
}

func (licenseServer *LicenseServer) getLatestVersion() (string, error) {
	listResponse, err := licenseServer.bucket.List(fmt.Sprintf("%s/", upgradePathPrefix), "/", "", 1000)
	if err != nil {
		return "", err
	}

	latestVersion := ""
	for _, key := range listResponse.Contents {
		if submatch := upgradeMatchRegex.FindStringSubmatch(key.Key); len(submatch) > 1 {
			version := submatch[1]
			if version > latestVersion {
				latestVersion = version
			}
		}
	}
	return latestVersion, nil
}

func (licenseServer *LicenseServer) printCheckUpgradeResponse(checkUpgradeResponse *license.CheckUpgradeResponse, writer http.ResponseWriter, request *http.Request) {
	jsonedCheckUpgradeResponse, err := json.Marshal(checkUpgradeResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedCheckUpgradeResponse)
}

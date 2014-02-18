package licenseserver

import (
	"encoding/json"
	"fmt"
	"io"
	"koality/license"
	"net/http"
)

func (licenseServer *LicenseServer) downloadUpgrade(writer http.ResponseWriter, request *http.Request) {
	downloadUpgradeRequest := new(license.DownloadUpgradeRequest)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(downloadUpgradeRequest); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if downloadUpgradeRequest.LicenseKey == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a license key")
		return
	} else if downloadUpgradeRequest.ServerId == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a server id")
		return
	} else if downloadUpgradeRequest.Version == "" {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a version to download")
		return
	}

	upgradeReader, err := licenseServer.bucket.GetReader(fmt.Sprintf("%s/koality-%s.tgz", upgradePathPrefix, downloadUpgradeRequest.Version))
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(writer, "No upgrade found with version %s", downloadUpgradeRequest.Version)
		return
	}

	defer upgradeReader.Close()
	bytesWritten, err := io.Copy(writer, upgradeReader)
	if bytesWritten == 0 && err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
}

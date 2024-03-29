package settings

import (
	"fmt"
	"koality/resources"
	"net/http"
)

func (settingsHandler *SettingsHandler) clearS3ExporterSettings(writer http.ResponseWriter, request *http.Request) {
	_, err := settingsHandler.resourcesConnection.Settings.Read.GetS3ExporterSettings()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		fmt.Fprintf(writer, "ok")
		return
	}

	err = settingsHandler.resourcesConnection.Settings.Delete.ClearS3ExporterSettings()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

func (settingsHandler *SettingsHandler) clearHipChatSettings(writer http.ResponseWriter, request *http.Request) {
	_, err := settingsHandler.resourcesConnection.Settings.Read.GetHipChatSettings()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		fmt.Fprintf(writer, "ok")
		return
	}

	err = settingsHandler.resourcesConnection.Settings.Delete.ClearHipChatSettings()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

func (settingsHandler *SettingsHandler) clearGitHubEnterpriseSettings(writer http.ResponseWriter, request *http.Request) {
	_, err := settingsHandler.resourcesConnection.Settings.Read.GetGitHubEnterpriseSettings()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		fmt.Fprintf(writer, "ok")
		return
	}

	err = settingsHandler.resourcesConnection.Settings.Delete.ClearGitHubEnterpriseSettings()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

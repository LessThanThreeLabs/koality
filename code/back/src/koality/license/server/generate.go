package licenseserver

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"koality/license"
	"net/http"
)

func (licenseServer *LicenseServer) generateLicense(writer http.ResponseWriter, request *http.Request) {
	licenseGenerateRequest := new(license.GenerateLicenseRequest)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(licenseGenerateRequest); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if licenseGenerateRequest.MaxExecutors == 0 {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Must provide a positive max executor cap")
		return
	}

	licenseKey, err := licenseServer.generateLicenseKey()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	id := uint64(0)
	query := "INSERT INTO licenses (key, max_executors) VALUES ($1, $2) RETURNING id"
	err = licenseServer.database.QueryRow(query, licenseKey, licenseGenerateRequest.MaxExecutors).Scan(&id)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	licenseGenerateResponse := license.GenerateLicenseResponse{
		LicenseKey: licenseKey,
	}

	jsonedLicenseGenerateResponse, err := json.Marshal(licenseGenerateResponse)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", jsonedLicenseGenerateResponse)
}

func (licenseServer *LicenseServer) generateLicenseKey() (string, error) {
	var apiKeyBuffer bytes.Buffer
	_, err := io.CopyN(&apiKeyBuffer, rand.Reader, 10)
	if err != nil {
		return "", err
	}

	// See http://en.wikipedia.org/wiki/Base32#Crockford.27s_Base32
	crockfordsBase32Alphabet := "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	licenseKey := base32.NewEncoding(crockfordsBase32Alphabet).EncodeToString(apiKeyBuffer.Bytes())
	return licenseKey, nil
}

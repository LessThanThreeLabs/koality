package licenseserver

import (
	"database/sql"
	"fmt"
	"github.com/LessThanThreeLabs/goamz/s3"
	"github.com/gorilla/mux"
	"koality/license"
	"net/http"
)

type LicenseServer struct {
	database *sql.DB
	bucket   *s3.Bucket
	address  string
}

func New(database *sql.DB, bucket *s3.Bucket, port uint16) *LicenseServer {
	address := fmt.Sprintf(":%d", port)
	return &LicenseServer{database, bucket, address}
}

func (licenseServer *LicenseServer) Start() error {
	router := licenseServer.createRouter()
	http.Handle("/", router)
	return http.ListenAndServe(licenseServer.address, nil)
}

func (licenseServer *LicenseServer) createRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc(license.PingRoute, func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "pong")
	}).Methods("GET")

	licenseSubRouter := router.PathPrefix(license.LicenseRoute).Subrouter()

	licenseSubRouter.HandleFunc(license.CheckLicenseSubroute, licenseServer.checkLicense).Methods("POST")
	licenseSubRouter.HandleFunc(license.GenerateLicenseSubroute, licenseServer.generateLicense).Methods("POST")

	licenseSubRouter.HandleFunc(license.DeactivateLicenseSubroute, licenseServer.deactivateLicense).Methods("PUT")
	licenseSubRouter.HandleFunc(license.ReactivateLicenseSubroute, licenseServer.reactivateLicense).Methods("PUT")
	licenseSubRouter.HandleFunc(license.SetLicenseMaxExecutorsSubroute, licenseServer.setMaxExecutors).Methods("PUT")

	upgradeSubRouter := router.PathPrefix(license.UpgradeRoute).Subrouter()

	upgradeSubRouter.HandleFunc(license.CheckUpgradeSubroute, licenseServer.checkUpgrade).Methods("POST")
	upgradeSubRouter.HandleFunc(license.DownloadUpgradeSubroute, licenseServer.downloadUpgrade).Methods("POST")

	return router
}

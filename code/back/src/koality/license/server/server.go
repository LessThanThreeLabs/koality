package licenseserver

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"koality/license"
	"net/http"
)

type LicenseServer struct {
	database *sql.DB
	address  string
}

func New(database *sql.DB, port uint16) *LicenseServer {
	address := fmt.Sprintf(":%d", port)
	return &LicenseServer{database, address}
}

func (licenseServer *LicenseServer) Start() error {
	router := licenseServer.createRouter()
	http.Handle("/", router)
	return http.ListenAndServe(licenseServer.address, nil)
}

func (licenseServer *LicenseServer) createRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc(license.CheckRoute, licenseServer.checkLicense).Methods("POST")
	router.HandleFunc(license.GenerateRoute, licenseServer.generateLicense).Methods("POST")

	router.HandleFunc(license.DeactivateRoute, licenseServer.deactivateLicense).Methods("PUT")
	router.HandleFunc(license.ReactivateRoute, licenseServer.reactivateLicense).Methods("PUT")
	router.HandleFunc(license.SetMaxExecutorsRoute, licenseServer.setMaxExecutors).Methods("PUT")

	return router
}

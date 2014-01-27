package webserver

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"koality/repositorymanager"
	"koality/resources"
	"koality/webserver/accounts"
	"koality/webserver/repositories"
	"koality/webserver/settings"
	"koality/webserver/stages"
	"koality/webserver/users"
	"koality/webserver/verifications"
	"net/http"
)

type Webserver struct {
	resourcesConnection *resources.Connection
	repositoryManager   repositorymanager.RepositoryManager
	sessionName         string
	address             string
}

func New(resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager, port int) (*Webserver, error) {
	address := fmt.Sprintf(":%d", port)
	return &Webserver{resourcesConnection, repositoryManager, "koality", address}, nil
}

func (webserver *Webserver) Start() error {
	sessionStore, err := webserver.createSessionStore()
	if err != nil {
		return err
	}

	router, err := webserver.createRouter(sessionStore)
	if err != nil {
		return err
	}

	loadUserIdRouter := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, _ := sessionStore.Get(request, webserver.sessionName)
		var _ = session
		// context.Set(request, "userId", session.Values["userId"])
		context.Set(request, "userId", uint64(1000))
		router.ServeHTTP(writer, request)
	})

	http.Handle("/", loadUserIdRouter)
	return http.ListenAndServe(webserver.address, nil)
}

func (webserver *Webserver) createSessionStore() (sessions.Store, error) {
	cookieStoreKeys, err := webserver.resourcesConnection.Settings.Read.GetCookieStoreKeys()
	if err != nil {
		return nil, err
	}

	sessionStore := sessions.NewCookieStore(cookieStoreKeys.Authentication, cookieStoreKeys.Encryption)
	// sessionStore.Options = &sessions.Options{
	// 	Path:     "/",
	// 	MaxAge:   2592000,
	// 	HttpOnly: true,
	// 	Secure:   true,
	// }
	return sessionStore, nil
}

func (webserver *Webserver) createRouter(sessionStore sessions.Store) (*mux.Router, error) {
	router := mux.NewRouter()

	passwordHasher, err := resources.NewPasswordHasher()
	if err != nil {
		return nil, err
	}

	accountsHandler, err := accounts.New(webserver.resourcesConnection, sessionStore, webserver.sessionName, passwordHasher)
	if err != nil {
		return nil, err
	}

	usersHandler, err := users.New(webserver.resourcesConnection, passwordHasher)
	if err != nil {
		return nil, err
	}

	repositoriesHandler, err := repositories.New(webserver.resourcesConnection, webserver.repositoryManager)
	if err != nil {
		return nil, err
	}

	verificationsHandler, err := verifications.New(webserver.resourcesConnection)
	if err != nil {
		return nil, err
	}

	stagesHandler, err := stages.New(webserver.resourcesConnection)
	if err != nil {
		return nil, err
	}

	settingsHandler, err := settings.New(webserver.resourcesConnection)
	if err != nil {
		return nil, err
	}

	apiSubrouter := router.PathPrefix("/api").MatcherFunc(webserver.hasApiKey).Subrouter()
	handleApiSubroute(apiSubrouter, webserver.resourcesConnection, webserver.repositoryManager)

	appSubrouter := router.PathPrefix("/app").Subrouter()
	accountsSubrouter := appSubrouter.PathPrefix("/accounts").Subrouter()
	accountsHandler.WireSubroutes(accountsSubrouter)

	usersSubrouter := appSubrouter.PathPrefix("/users").MatcherFunc(webserver.isLoggedIn).Subrouter()
	usersHandler.WireSubroutes(usersSubrouter)

	repositoriesSubrouter := appSubrouter.PathPrefix("/repositories").MatcherFunc(webserver.isLoggedIn).Subrouter()
	repositoriesHandler.WireSubroutes(repositoriesSubrouter)

	verificationsSubrouter := appSubrouter.PathPrefix("/verifications").MatcherFunc(webserver.isLoggedIn).Subrouter()
	verificationsHandler.WireSubroutes(verificationsSubrouter)

	stagesSubrouter := appSubrouter.PathPrefix("/stages").MatcherFunc(webserver.isLoggedIn).Subrouter()
	stagesHandler.WireStagesSubroutes(stagesSubrouter)

	stageRunsSubrouter := appSubrouter.PathPrefix("/stageRuns").MatcherFunc(webserver.isLoggedIn).Subrouter()
	stagesHandler.WireStageRunsSubroutes(stageRunsSubrouter)

	settingsSubrouter := appSubrouter.PathPrefix("/settings").MatcherFunc(webserver.isLoggedIn).Subrouter()
	settingsHandler.WireSubroutes(settingsSubrouter)

	return router, nil
}

func (webserver *Webserver) isLoggedIn(request *http.Request, match *mux.RouteMatch) bool {
	userId, ok := context.Get(request, "userId").(uint64)
	if !ok {
		return false
	}
	return userId != 0
}

func (webserver *Webserver) hasApiKey(request *http.Request, match *mux.RouteMatch) bool {
	apiKeyToCheck := request.FormValue("key")
	apiKey := "need-to-actually-get-this-from-the-database"
	// apiKey, err := webserver.resourcesConnection.Settings.Read.GetApiKey()
	// if err != nil {
	// 	fmt.Println("webserver - hasApiKey: NEED TO LOG THIS")
	// 	return false
	// }
	return apiKey == apiKeyToCheck
}

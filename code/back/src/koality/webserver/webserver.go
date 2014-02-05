package webserver

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	githubconnection "koality/github"
	"koality/repositorymanager"
	"koality/resources"
	"koality/webserver/accounts"
	"koality/webserver/github"
	"koality/webserver/middleware"
	"koality/webserver/repositories"
	"koality/webserver/settings"
	"koality/webserver/stages"
	"koality/webserver/templates"
	"koality/webserver/users"
	"koality/webserver/verifications"
	"net/http"
	"strings"
)

type Webserver struct {
	resourcesConnection *resources.Connection
	repositoryManager   repositorymanager.RepositoryManager
	gitHubConnection    githubconnection.GitHubConnection
	sessionName         string
	address             string
}

func New(resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager, gitHubConnection githubconnection.GitHubConnection, port int) (*Webserver, error) {
	address := fmt.Sprintf(":%d", port)
	return &Webserver{resourcesConnection, repositoryManager, gitHubConnection, "koality", address}, nil
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

	hasCsrfTokenWrapper := middleware.CheckCsrfTokenWraper(sessionStore, webserver.sessionName, router)
	hasApiKeyWrapper := middleware.HasApiKeyWrapper(webserver.resourcesConnection, router)

	loadUserIdRouter := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, _ := sessionStore.Get(request, webserver.sessionName)
		// set to zero if can't find!!
		// context.Set(request, "userId", session.Values["userId"])
		var _ = session
		context.Set(request, "userId", uint64(1000))
		// context.Set(request, "userId", uint64(0))

		if request.URL.Path == "/" || request.URL.Path == "/dashboard" ||
			strings.HasPrefix(request.URL.Path, "/repository/") ||
			strings.HasPrefix(request.URL.Path, "/hooks/") ||
			strings.HasPrefix(request.URL.Path, "/oAuth/") {
			router.ServeHTTP(writer, request)
		} else if strings.HasPrefix(request.URL.Path, "/app/") {
			hasCsrfTokenWrapper(writer, request)
		} else if strings.HasPrefix(request.URL.Path, "/api/") {
			hasApiKeyWrapper(writer, request)
		} else {
			panic("Unexpected path: " + request.URL.Path)
		}
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
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}
	return sessionStore, nil
}

func (webserver *Webserver) createRouter(sessionStore sessions.Store) (*mux.Router, error) {
	router := mux.NewRouter()

	passwordHasher, err := resources.NewPasswordHasher()
	if err != nil {
		return nil, err
	}

	templatesHandler, err := templates.New(webserver.resourcesConnection, sessionStore, webserver.sessionName)
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

	repositoriesHandler, err := repositories.New(webserver.resourcesConnection, webserver.repositoryManager, webserver.gitHubConnection)
	if err != nil {
		return nil, err
	}

	verificationsHandler, err := verifications.New(webserver.resourcesConnection, webserver.repositoryManager)
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

	gitHubHandler, err := github.New(webserver.resourcesConnection, webserver.repositoryManager)
	if err != nil {
		return nil, err
	}

	wireRootSubroutes := func() {
		templatesHandler.WireRootSubroutes(router)
	}

	wireAppSubroutes := func() {
		appSubrouter := router.PathPrefix("/app").Subrouter()
		accountsSubrouter := appSubrouter.PathPrefix("/accounts").Subrouter()
		accountsHandler.WireAppSubroutes(accountsSubrouter)

		usersSubrouter := appSubrouter.PathPrefix("/users").Subrouter()
		usersHandler.WireAppSubroutes(usersSubrouter)

		repositoriesSubrouter := appSubrouter.PathPrefix("/repositories").Subrouter()
		repositoriesHandler.WireAppSubroutes(repositoriesSubrouter)

		verificationsSubrouter := appSubrouter.PathPrefix("/verifications").Subrouter()
		verificationsHandler.WireAppSubroutes(verificationsSubrouter)

		stagesSubrouter := appSubrouter.PathPrefix("/stages").Subrouter()
		stagesHandler.WireStagesAppSubroutes(stagesSubrouter)

		stageRunsSubrouter := appSubrouter.PathPrefix("/stageRuns").Subrouter()
		stagesHandler.WireStageRunsAppSubroutes(stageRunsSubrouter)

		settingsSubrouter := appSubrouter.PathPrefix("/settings").Subrouter()
		settingsHandler.WireAppSubroutes(settingsSubrouter)
	}

	wireApiSubroutes := func() {
		apiSubrouter := router.PathPrefix("/api").Subrouter()

		userSubrouter := apiSubrouter.PathPrefix("/users").Subrouter()
		usersHandler.WireApiSubroutes(userSubrouter)

		repositoriesSubrouter := apiSubrouter.PathPrefix("/repositories").Subrouter()
		repositoriesHandler.WireApiSubroutes(repositoriesSubrouter)

		verificationsSubrouter := apiSubrouter.PathPrefix("/verifications").Subrouter()
		verificationsHandler.WireApiSubroutes(verificationsSubrouter)

		stagesSubrouter := apiSubrouter.PathPrefix("/stages").Subrouter()
		stagesHandler.WireStagesApiSubroutes(stagesSubrouter)

		stageRunsSubrouter := apiSubrouter.PathPrefix("/stageRuns").Subrouter()
		stagesHandler.WireStageRunsApiSubroutes(stageRunsSubrouter)

		settingsSubrouter := apiSubrouter.PathPrefix("/settings").Subrouter()
		settingsHandler.WireApiSubroutes(settingsSubrouter)
	}

	wireHooksSubroutes := func() {
		hooksSubrouter := router.PathPrefix("/hooks").Subrouter()

		gitHubSubrouter := hooksSubrouter.PathPrefix("/gitHub").Subrouter()
		gitHubHandler.WireHooksSubroutes(gitHubSubrouter)
	}

	wireOAuthSubroutes := func() {
		oAuthSubrouter := router.PathPrefix("/oAuth").Subrouter()

		gitHubSubrouter := oAuthSubrouter.PathPrefix("/gitHub").Subrouter()
		gitHubHandler.WireOAuthSubroutes(gitHubSubrouter)
	}

	wireRootSubroutes()
	wireAppSubroutes()
	wireApiSubroutes()
	wireHooksSubroutes()
	wireOAuthSubroutes()

	return router, nil
}

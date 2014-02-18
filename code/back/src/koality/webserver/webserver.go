package webserver

import (
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	githubconnection "koality/github"
	"koality/license/client"
	"koality/mail"
	"koality/repositorymanager"
	"koality/resources"
	"koality/webserver/accounts"
	"koality/webserver/builds"
	"koality/webserver/events"
	"koality/webserver/feedback"
	"koality/webserver/github"
	"koality/webserver/google"
	"koality/webserver/middleware"
	"koality/webserver/repositories"
	"koality/webserver/settings"
	"koality/webserver/stages"
	"koality/webserver/templates"
	"koality/webserver/users"
	"koality/webserver/websockets"
	"net/http"
	"strings"
)

type Webserver struct {
	resourcesConnection *resources.Connection
	repositoryManager   repositorymanager.RepositoryManager
	gitHubConnection    githubconnection.GitHubConnection
	mailer              mail.Mailer
	licenseManager      *licenseclient.LicenseManager
	sessionName         string
	address             string
}

func New(resourcesConnection *resources.Connection, repositoryManager repositorymanager.RepositoryManager, gitHubConnection githubconnection.GitHubConnection, mailer mail.Mailer, licenseManager *licenseclient.LicenseManager, port uint16) (*Webserver, error) {
	address := fmt.Sprintf(":%d", port)
	return &Webserver{resourcesConnection, repositoryManager, gitHubConnection, mailer, licenseManager, "koality", address}, nil
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

	hasHeaderCsrfTokenWrapper := middleware.CheckCsrfTokenWrapper(sessionStore, webserver.sessionName, router, false)
	hasQueryCsrfTokenWrapper := middleware.CheckCsrfTokenWrapper(sessionStore, webserver.sessionName, router, true)
	hasApiKeyWrapper := middleware.HasApiKeyWrapper(webserver.resourcesConnection, router)

	loadUserIdRouter := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, _ := sessionStore.Get(request, webserver.sessionName)
		userId, ok := session.Values["userId"]
		if !ok {
			userId = uint64(0)
		}
		context.Set(request, "userId", userId)

		if strings.HasPrefix(request.URL.Path, "/websockets/") {
			router.ServeHTTP(writer, request)
			hasQueryCsrfTokenWrapper(writer, request)
		} else if strings.HasPrefix(request.URL.Path, "/app/") {
			hasHeaderCsrfTokenWrapper(writer, request)
		} else if strings.HasPrefix(request.URL.Path, "/api/") {
			hasApiKeyWrapper(writer, request)
		} else {
			router.ServeHTTP(writer, request)
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
		MaxAge:   0,
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

	websocketsManager, err := websockets.NewManager()
	if err != nil {
		return nil, err
	}

	templatesHandler, err := templates.New(webserver.resourcesConnection, sessionStore, webserver.sessionName)
	if err != nil {
		return nil, err
	}

	websocketsHandler, err := websockets.New(websocketsManager)
	if err != nil {
		return nil, err
	}

	eventsHandler, err := events.New(webserver.resourcesConnection, websocketsManager)
	if err != nil {
		return nil, err
	}

	accountsHandler, err := accounts.New(webserver.resourcesConnection, sessionStore, webserver.sessionName, passwordHasher, webserver.mailer)
	if err != nil {
		return nil, err
	}

	usersHandler, err := users.New(webserver.resourcesConnection, passwordHasher, webserver.gitHubConnection)
	if err != nil {
		return nil, err
	}

	repositoriesHandler, err := repositories.New(webserver.resourcesConnection, webserver.repositoryManager, webserver.gitHubConnection)
	if err != nil {
		return nil, err
	}

	buildsHandler, err := builds.New(webserver.resourcesConnection, webserver.repositoryManager)
	if err != nil {
		return nil, err
	}

	stagesHandler, err := stages.New(webserver.resourcesConnection)
	if err != nil {
		return nil, err
	}

	settingsHandler, err := settings.New(webserver.resourcesConnection, sessionStore, webserver.sessionName, passwordHasher, webserver.licenseManager)
	if err != nil {
		return nil, err
	}

	gitHubHandler, err := github.New(webserver.resourcesConnection, webserver.repositoryManager, webserver.gitHubConnection)
	if err != nil {
		return nil, err
	}

	googleHandler, err := google.New(webserver.resourcesConnection, sessionStore, webserver.sessionName)
	if err != nil {
		return nil, err
	}

	feedbackHandler, err := feedback.New(webserver.resourcesConnection, webserver.mailer)
	if err != nil {
		return nil, err
	}

	wireTemplateSubroutes := func() {
		templatesHandler.WireTemplateSubroutes(router)
	}

	wireWebsocketSubroutes := func() {
		websocketSubrouter := router.PathPrefix("/websockets").Subrouter()
		websocketsHandler.WireWebsocketSubroutes(websocketSubrouter)
	}

	wireAppSubroutes := func() {
		appSubrouter := router.PathPrefix("/app").Subrouter()

		eventsSubrouter := appSubrouter.PathPrefix("/events").Subrouter()
		eventsHandler.WireAppSubroutes(eventsSubrouter)

		accountsSubrouter := appSubrouter.PathPrefix("/accounts").Subrouter()
		accountsHandler.WireAppSubroutes(accountsSubrouter)

		usersSubrouter := appSubrouter.PathPrefix("/users").Subrouter()
		usersHandler.WireAppSubroutes(usersSubrouter)

		repositoriesSubrouter := appSubrouter.PathPrefix("/repositories").Subrouter()
		repositoriesHandler.WireAppSubroutes(repositoriesSubrouter)

		buildsSubrouter := appSubrouter.PathPrefix("/builds").Subrouter()
		buildsHandler.WireAppSubroutes(buildsSubrouter)

		stagesSubrouter := appSubrouter.PathPrefix("/stages").Subrouter()
		stagesHandler.WireStagesAppSubroutes(stagesSubrouter)

		stageRunsSubrouter := appSubrouter.PathPrefix("/stageRuns").Subrouter()
		stagesHandler.WireStageRunsAppSubroutes(stageRunsSubrouter)

		settingsSubrouter := appSubrouter.PathPrefix("/settings").Subrouter()
		settingsHandler.WireAppSubroutes(settingsSubrouter)

		feedbackSubrouter := appSubrouter.PathPrefix("/feedback").Subrouter()
		feedbackHandler.WireAppSubroutes(feedbackSubrouter)
	}

	wireApiSubroutes := func() {
		apiSubrouter := router.PathPrefix("/api").Subrouter()

		userSubrouter := apiSubrouter.PathPrefix("/users").Subrouter()
		usersHandler.WireApiSubroutes(userSubrouter)

		repositoriesSubrouter := apiSubrouter.PathPrefix("/repositories").Subrouter()
		repositoriesHandler.WireApiSubroutes(repositoriesSubrouter)

		buildsSubrouter := apiSubrouter.PathPrefix("/builds").Subrouter()
		buildsHandler.WireApiSubroutes(buildsSubrouter)

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

		googleSubrouter := oAuthSubrouter.PathPrefix("/google").Subrouter()
		googleHandler.WireOAuthSubroutes(googleSubrouter)
	}

	wirePing := func() {
		router.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprint(writer, "pong")
		}).Methods("GET")
	}

	wireTemplateSubroutes()
	wireWebsocketSubroutes()
	wireAppSubroutes()
	wireApiSubroutes()
	wireHooksSubroutes()
	wireOAuthSubroutes()
	wirePing()

	return router, nil
}

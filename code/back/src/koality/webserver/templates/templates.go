package templates

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"io"
	"io/ioutil"
	"koality/resources"
	"koality/util/pathtranslator"
	"net/http"
	"path"
)

type indexTemplateValues struct {
	User                   *resources.User
	CsrfToken              string
	AuthenticationSettings *resources.AuthenticationSettings
	CssFiles               []string
	JsFiles                []string
}

type wizardTemplateValues struct {
	CsrfToken string
	CssFiles  []string
}

type TemplatesHandler struct {
	resourcesConnection *resources.Connection
	sessionStore        sessions.Store
	sessionName         string
	indexTemplate       *template.Template
	wizardTemplate      *template.Template
	cssFiles            []string
	jsFiles             []string
}

var (
	pathsToHandle []string = []string{"index", "login", "create/account", "account", "resetPassword", "admin", "dashboard", "repository/{repositoryId:[0-9]+}"}
)

func New(resourcesConnection *resources.Connection, sessionStore sessions.Store, sessionName string) (*TemplatesHandler, error) {
	indexTemplate, err := getIndexTemplate()
	if err != nil {
		return nil, err
	}

	wizardTemplate, err := getWizardTemplate()
	if err != nil {
		return nil, err
	}

	cssFiles, jsFiles, err := getCssAndJsFiles()
	if err != nil {
		panic(err)
	}
	return &TemplatesHandler{resourcesConnection, sessionStore, sessionName, indexTemplate, wizardTemplate, cssFiles, jsFiles}, nil
}

func getIndexTemplate() (*template.Template, error) {
	relativePath := path.Join("code", "front", "templates", "index.html")
	filePath, err := pathtranslator.TranslatePathAndCheckExists(relativePath)
	if err != nil {
		return nil, err
	}
	return template.ParseFiles(filePath)
}

func getWizardTemplate() (*template.Template, error) {
	relativePath := path.Join("code", "front", "templates", "wizard.html")
	filePath, err := pathtranslator.TranslatePathAndCheckExists(relativePath)
	if err != nil {
		return nil, err
	}
	return template.ParseFiles(filePath)
}

func getCssAndJsFiles() ([]string, []string, error) {
	relativePath := path.Join("code", "front", "templates", "indexFiles.json")
	filePath, err := pathtranslator.TranslatePathAndCheckExists(relativePath)
	if err != nil {
		return nil, nil, err
	}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	var cssAndJsFiles map[string][]string
	err = json.Unmarshal(file, &cssAndJsFiles)
	if err != nil {
		return nil, nil, err
	}
	return cssAndJsFiles["css"], cssAndJsFiles["js"], nil
}

func (templatesHandler *TemplatesHandler) WireTemplateSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/", templatesHandler.getRoot).Methods("GET")
	for _, pathToHandle := range pathsToHandle {
		subrouter.HandleFunc(fmt.Sprintf("/%s", pathToHandle), templatesHandler.getRoot).Methods("GET")
		subrouter.HandleFunc(fmt.Sprintf("/%s.html", pathToHandle), templatesHandler.getRoot).Methods("GET")
	}

	subrouter.HandleFunc("/wizard", templatesHandler.getWizard).Methods("GET")
	subrouter.HandleFunc("/wizard.html", templatesHandler.getWizard).Methods("GET")
}

func (templatesHandler *TemplatesHandler) getRoot(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)
	user, err := templatesHandler.resourcesConnection.Users.Read.Get(userId)
	if _, ok := err.(resources.NoSuchUserError); err != nil && !ok {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	csrfToken, err := templatesHandler.getCsrfFromSession(writer, request)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	authenticationSettings, err := templatesHandler.resourcesConnection.Settings.Read.GetAuthenticationSettings()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	templateValues := indexTemplateValues{user, csrfToken, authenticationSettings, templatesHandler.cssFiles, templatesHandler.jsFiles}
	templatesHandler.indexTemplate.Execute(writer, templateValues)
}

func (templatesHandler *TemplatesHandler) getWizard(writer http.ResponseWriter, request *http.Request) {
	csrfToken, err := templatesHandler.getCsrfFromSession(writer, request)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	templateValues := wizardTemplateValues{csrfToken, templatesHandler.cssFiles}
	templatesHandler.wizardTemplate.Execute(writer, templateValues)
}

func (templatesHandler *TemplatesHandler) getCsrfFromSession(writer http.ResponseWriter, request *http.Request) (string, error) {
	session, _ := templatesHandler.sessionStore.Get(request, templatesHandler.sessionName)
	csrfToken, ok := session.Values["csrfToken"]
	if ok {
		return csrfToken.(string), nil
	} else {
		var csrfTokenBuffer bytes.Buffer
		if _, err := io.CopyN(&csrfTokenBuffer, rand.Reader, 15); err != nil {
			return "", err
		}
		newCsrfToken := base32.StdEncoding.EncodeToString(csrfTokenBuffer.Bytes())
		session.Values["csrfToken"] = newCsrfToken
		session.Save(request, writer)
		return newCsrfToken, nil
	}
}

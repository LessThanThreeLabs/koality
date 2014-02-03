package templates

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"html/template"
	"io"
	"koality/resources"
	"koality/util/pathtranslator"
	"net/http"
	"path"
)

type indexTemplateValues struct {
	User      *resources.User
	CsrfToken string
}

type TemplatesHandler struct {
	resourcesConnection *resources.Connection
	sessionStore        sessions.Store
	sessionName         string
	indexTemplate       *template.Template
}

func New(resourcesConnection *resources.Connection, sessionStore sessions.Store, sessionName string) (*TemplatesHandler, error) {
	indexTemplate, err := getIndexTemplate()
	if err != nil {
		return nil, err
	}
	return &TemplatesHandler{resourcesConnection, sessionStore, sessionName, indexTemplate}, nil
}

func getIndexTemplate() (*template.Template, error) {
	relativePath := path.Join("code", "front", "templates", "index.html")
	filePath, err := pathtranslator.TranslatePathAndCheckExists(relativePath)
	if err != nil {
		return nil, err
	}
	return template.ParseFiles(filePath)
}

func (templatesHandler *TemplatesHandler) WireRootSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/", templatesHandler.getRoot).Methods("GET")
	subrouter.HandleFunc("/index", templatesHandler.getRoot).Methods("GET")
	subrouter.HandleFunc("/index.html", templatesHandler.getRoot).Methods("GET")
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

	templateValues := indexTemplateValues{user, csrfToken}
	templatesHandler.indexTemplate.Execute(writer, templateValues)
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

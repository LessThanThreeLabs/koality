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
	UserId    uint64
	IsAdmin   bool
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
	subrouter.HandleFunc("/", templatesHandler.GetRoot).Methods("GET")
	subrouter.HandleFunc("/index", templatesHandler.GetRoot).Methods("GET")
	subrouter.HandleFunc("/index.html", templatesHandler.GetRoot).Methods("GET")
}

func (templatesHandler *TemplatesHandler) GetRoot(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)
	user, err := templatesHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	var csrfTokenBuffer bytes.Buffer
	_, err = io.CopyN(&csrfTokenBuffer, rand.Reader, 15)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to stringify: %v", err)
		return
	}
	csrfToken := base32.StdEncoding.EncodeToString(csrfTokenBuffer.Bytes())

	session, _ := templatesHandler.sessionStore.Get(request, templatesHandler.sessionName)
	session.Values["csrfToken"] = csrfToken
	session.Save(request, writer)

	templateValues := indexTemplateValues{userId, user.IsAdmin, csrfToken}
	templatesHandler.indexTemplate.Execute(writer, templateValues)
}

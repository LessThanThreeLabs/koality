package feedback

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"html"
	"koality/mail"
	"koality/resources"
	"koality/webserver/middleware"
	"net/http"
	"strings"
)

type FeedbackRequestData struct {
	Feedback     string `json:"feedback"`
	UserAgent    string `json:"userAgent"`
	WindowWidth  int    `json:"windowWidth"`
	WindowHeight int    `json:"windowHeight"`
}

type FeedbackHandler struct {
	resourcesConnection *resources.Connection
	mailer              mail.Mailer
}

func New(resourcesConnection *resources.Connection, mailer mail.Mailer) (*FeedbackHandler, error) {
	return &FeedbackHandler{resourcesConnection, mailer}, nil
}

func (feedbackHandler *FeedbackHandler) WireAppSubroutes(subrouter *mux.Router) {
	subrouter.HandleFunc("/",
		middleware.IsLoggedInWrapper(feedbackHandler.sendFeedback)).
		Methods("POST")
}

func (feedbackHandler *FeedbackHandler) sendFeedback(writer http.ResponseWriter, request *http.Request) {
	feedbackRequestData := new(FeedbackRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(feedbackRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	userId := context.Get(request, "userId").(uint64)
	user, err := feedbackHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	domainName, err := feedbackHandler.resourcesConnection.Settings.Read.GetDomainName()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		domainName = "uninitialized-koality"
	}

	message := fmt.Sprintf("User: %s %s (%s)\n\nFeedback: %s\n\nUser Agent: %s\n\nWindow: %d x %d",
		user.FirstName, user.LastName, user.Email, feedbackRequestData.Feedback, feedbackRequestData.UserAgent,
		feedbackRequestData.WindowWidth, feedbackRequestData.WindowHeight)
	message = strings.Replace(html.EscapeString(message), "\n", "<br>", -1)
	replyTo := []string{user.Email, "feedback@koalitycode.com"}
	to := []string{"feedback@koalitycode.com"}

	err = feedbackHandler.mailer.SendMail(fmt.Sprintf("feedback@%s", domainName), replyTo, to, "Feedback", message)
	if _, ok := err.(mail.NoAuthProvidedError); ok {
		var errorMessage string
		if user.IsAdmin {
			errorMessage = "Unable to send feedback email, please configure SMTP settings first"
		} else {
			errorMessage = "Unable to send feedback email, an Administrator must configure SMTP settings first"
		}
		writer.WriteHeader(http.StatusPreconditionFailed)
		fmt.Fprint(writer, errorMessage)
		return
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "ok")
}

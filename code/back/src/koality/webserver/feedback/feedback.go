package feedback

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"koality/mail"
	"koality/resources"
	"koality/webserver/middleware"
	"net/http"
)

const (
	rememberMeDuration = 2592000
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

	message := fmt.Sprintf("User: %s %s (%s)\n\nFeedback: %s\n\nUser Agent: %s\n\nWindow: %d x %d",
		user.FirstName, user.LastName, user.Email, feedbackRequestData.Feedback, feedbackRequestData.UserAgent,
		feedbackRequestData.WindowWidth, feedbackRequestData.WindowHeight)
	domainName := "thisShouldBeARealDomainName"
	userEmail := "thisShouldBeARealUserEmailAddress"
	replyTo := []string{userEmail, "feedback@koalitycode.com"}
	to := []string{"feedback@koalitycode.com"}
	if err = feedbackHandler.mailer.SendMail("feedback@"+domainName, replyTo, to, "Feedback", message); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "ok")
}

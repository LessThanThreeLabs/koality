package accounts

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"koality/mail"
	"koality/resources"
	"koality/webserver/util"
	"net/http"
)

func (accountsHandler *AccountsHandler) login(writer http.ResponseWriter, request *http.Request) {
	loginRequestData := new(loginRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(loginRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	user, err := accountsHandler.resourcesConnection.Users.Read.GetByEmail(loginRequestData.Email)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	passwordHash, err := accountsHandler.passwordHasher.ComputeHash(loginRequestData.Password, user.PasswordSalt)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if bytes.Compare(passwordHash, user.PasswordHash) != 0 {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Invalid password")
		return
	}

	session, _ := accountsHandler.sessionStore.Get(request, accountsHandler.sessionName)
	util.Login(user.Id, loginRequestData.RememberMe, session, writer, request)

	fmt.Fprint(writer, "ok")
}

func (accountsHandler *AccountsHandler) logout(writer http.ResponseWriter, request *http.Request) {
	session, _ := accountsHandler.sessionStore.Get(request, accountsHandler.sessionName)
	util.Logout(session, writer, request)

	fmt.Fprint(writer, "ok")
}

func (accountsHandler *AccountsHandler) resetPassword(writer http.ResponseWriter, request *http.Request) {
	emailBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	email := string(emailBytes)

	user, err := accountsHandler.resourcesConnection.Users.Read.GetByEmail(email)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		fmt.Fprint(writer, "No user found with email: "+email)
		return
	}

	domainName, err := accountsHandler.resourcesConnection.Settings.Read.GetDomainName()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		domainName = "koalitycode.com"
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	var passwordBuffer bytes.Buffer
	if _, err = io.CopyN(&passwordBuffer, rand.Reader, 10); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	// See http://en.wikipedia.org/wiki/Base32#Crockford.27s_Base32
	crockfordsBase32Alphabet := "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	password := base32.NewEncoding(crockfordsBase32Alphabet).EncodeToString(passwordBuffer.Bytes())

	passwordHash, passwordSalt, err := accountsHandler.passwordHasher.GenerateHashAndSalt(password)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	undoPasswordReset := func() {
		accountsHandler.resourcesConnection.Users.Update.SetPassword(user.Id, user.PasswordHash, user.PasswordSalt)
	}

	if err = accountsHandler.resourcesConnection.Users.Update.SetPassword(user.Id, passwordHash, passwordSalt); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fromAddress := fmt.Sprintf("koality@%s", domainName)
	replyToAddresses := []string{fmt.Sprintf("noreply@%s", domainName)}
	toAddresses := []string{email}
	emailBody := fmt.Sprintf("Your new password is: %s", password)

	err = accountsHandler.mailer.SendMail(fromAddress, replyToAddresses, toAddresses, "Your new Koality Password", emailBody)
	if _, ok := err.(mail.NoAuthProvidedError); ok {
		undoPasswordReset()
		writer.WriteHeader(http.StatusPreconditionFailed)
		fmt.Fprint(writer, "Unable to send confirmation email, an Administrator must configure SMTP settings first")
		return
	} else if err != nil {
		undoPasswordReset()
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "ok")
}

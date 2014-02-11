package accounts

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"koality/resources"
	"net/http"
)

func (accountsHandler *AccountsHandler) getMaxSessionAge(rememberMe bool) int {
	if rememberMe {
		return rememberMeDuration
	} else {
		return 0
	}
}

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
	session.Values["userId"] = user.Id
	session.Options.MaxAge = accountsHandler.getMaxSessionAge(loginRequestData.RememberMe)
	session.Save(request, writer)

	fmt.Fprint(writer, "ok")
}

func (accountsHandler *AccountsHandler) logout(writer http.ResponseWriter, request *http.Request) {
	session, _ := accountsHandler.sessionStore.Get(request, accountsHandler.sessionName)
	session.Options.MaxAge = -1
	session.Save(request, writer)

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

	if err = accountsHandler.resourcesConnection.Users.Update.SetPassword(user.Id, passwordHash, passwordSalt); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fromAddress := fmt.Sprintf("koality@%s", domainName)
	replyToAddresses := []string{fmt.Sprintf("noreply@%s", domainName)}
	toAddresses := []string{email}
	emailBody := fmt.Sprintf("Your new password is: %s", password)

	if err = accountsHandler.mailer.SendMail(fromAddress, replyToAddresses, toAddresses, "Your new Koality Password", emailBody); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "need to reset password for "+string(emailBytes))
}

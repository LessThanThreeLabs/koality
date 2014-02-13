package accounts

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"koality/mail"
	"koality/resources"
	"koality/webserver/util"
	"net/http"
	"net/url"
)

func (accountsHandler *AccountsHandler) create(writer http.ResponseWriter, request *http.Request) {
	authenticationSettings, err := accountsHandler.resourcesConnection.Settings.Read.GetAuthenticationSettings()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if !authenticationSettings.ManualAccountsAllowed {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, "The administrator has disabled manual account creation")
		return
	}

	createAccountData := new(createAccountRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(createAccountData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Println(createAccountData)

	if err = accountsHandler.resourcesConnection.Users.Create.CanCreate(createAccountData.Email, createAccountData.FirstName, createAccountData.LastName); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	passwordHash, passwordSalt, err := accountsHandler.passwordHasher.GenerateHashAndSalt(createAccountData.Password)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	createAccountTokenBytes := make([]byte, 15)

	if _, err = rand.Read(createAccountTokenBytes); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	// See http://en.wikipedia.org/wiki/Base32#Crockford.27s_Base32
	crockfordsBase32Alphabet := "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	createAccountToken := base32.NewEncoding(crockfordsBase32Alphabet).EncodeToString(createAccountTokenBytes)

	confirmAccountData := confirmAccountData{createAccountData.Email, createAccountData.FirstName, createAccountData.LastName, passwordHash, passwordSalt, createAccountToken}

	session, _ := accountsHandler.sessionStore.Get(request, accountsHandler.sessionName)
	// session.Flashes()
	session.AddFlash(confirmAccountData)
	session.Save(request, writer)

	domainName, err := accountsHandler.resourcesConnection.Settings.Read.GetDomainName()
	if _, ok := err.(resources.NoSuchSettingError); ok {
		domainName = "koalitycode.com"
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fromAddress := fmt.Sprintf("koality@%s", domainName)
	replyToAddresses := []string{fmt.Sprintf("noreply@%s", domainName)}
	toAddresses := []string{createAccountData.Email}
	emailSubject := "Confirm your Koality account"
	confirmAccountQueryValues := url.Values{}
	confirmAccountQueryValues.Add("token", createAccountToken)
	confirmAccountUri := fmt.Sprintf("https://%s/create/account?%s", domainName, confirmAccountQueryValues.Encode())
	emailBody := fmt.Sprintf("Click on the following link to confirm your Koality account: <a href=\"%s\">%s</a>", confirmAccountUri, confirmAccountUri)

	err = accountsHandler.mailer.SendMail(fromAddress, replyToAddresses, toAddresses, emailSubject, emailBody)
	if _, ok := err.(mail.NoAuthProvidedError); ok {
		writer.WriteHeader(http.StatusPreconditionFailed)
		fmt.Fprint(writer, "Unable to send confirmation email, an Administrator must configure SMTP settings first")
		return
	} else if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	fmt.Fprint(writer, "ok")
}

func (accountsHandler *AccountsHandler) confirm(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	confirmAccountTokenBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	confirmAccountToken := string(confirmAccountTokenBytes)

	session, _ := accountsHandler.sessionStore.Get(request, accountsHandler.sessionName)
	confirmAccountFlashes := session.Flashes()

	var createAccountData *confirmAccountData
	for _, confirmAccountFlash := range confirmAccountFlashes {
		confirmAccountFlash, ok := confirmAccountFlash.(*confirmAccountData)
		if ok && confirmAccountFlash.Token == confirmAccountToken {
			createAccountData = confirmAccountFlash
		}
	}

	if createAccountData == nil {
		writer.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(writer, "Invalid account confirmation token")
		return
	}

	user, err := accountsHandler.resourcesConnection.Users.Create.Create(createAccountData.Email, createAccountData.FirstName, createAccountData.LastName, createAccountData.PasswordHash, createAccountData.PasswordSalt, false)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	util.Login(user.Id, false, session, writer, request)

	fmt.Fprint(writer, "ok")
}

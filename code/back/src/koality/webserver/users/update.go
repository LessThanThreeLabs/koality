package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (usersHandler *UsersHandler) setName(writer http.ResponseWriter, request *http.Request) {
	setNameRequestData := new(setNameRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(setNameRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	userId := context.Get(request, "userId").(uint64)
	err := usersHandler.resourcesConnection.Users.Update.SetName(userId, setNameRequestData.FirstName, setNameRequestData.LastName)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

func (usersHandler *UsersHandler) setPassword(writer http.ResponseWriter, request *http.Request) {
	setPasswordRequestData := new(setPasswordRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(setPasswordRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	userId := context.Get(request, "userId").(uint64)
	user, err := usersHandler.resourcesConnection.Users.Read.Get(userId)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	passwordHashToCheck, err := usersHandler.passwordHasher.ComputeHash(setPasswordRequestData.OldPassword, user.PasswordSalt)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	if bytes.Compare(user.PasswordHash, passwordHashToCheck) != 0 {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, invalid old password")
		return
	}

	passwordHash, passwordSalt, err := usersHandler.passwordHasher.GenerateHashAndSalt(setPasswordRequestData.NewPassword)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	err = usersHandler.resourcesConnection.Users.Update.SetPassword(userId, passwordHash, passwordSalt)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

func (usersHandler *UsersHandler) setAdmin(writer http.ResponseWriter, request *http.Request) {
	userId := context.Get(request, "userId").(uint64)

	userToModifyIdString := mux.Vars(request)["userId"]
	userToModifyId, err := strconv.ParseUint(userToModifyIdString, 10, 64)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse userId: %v", err)
		return
	}

	adminStatusString := request.PostFormValue("admin")
	adminStatus, err := strconv.ParseBool(adminStatusString)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Unable to parse admin: %v", err)
		return
	}

	if userId == userToModifyId {
		writer.WriteHeader(http.StatusForbidden)
		fmt.Fprint(writer, "Forbidden request, cannot modify admin status for self")
		return
	}

	err = usersHandler.resourcesConnection.Users.Update.SetAdmin(userToModifyId, adminStatus)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

func (usersHandler *UsersHandler) addKey(writer http.ResponseWriter, request *http.Request) {
	addKeyRequestData := new(addKeyRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(addKeyRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	userId := context.Get(request, "userId").(uint64)
	keyId, err := usersHandler.resourcesConnection.Users.Update.AddKey(userId, addKeyRequestData.Name, addKeyRequestData.PublicKey)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, keyId)
}

func (usersHandler *UsersHandler) removeKey(writer http.ResponseWriter, request *http.Request) {
	removeKeyRequestData := new(removeKeyRequestData)
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(removeKeyRequestData); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}

	userId := context.Get(request, "userId").(uint64)
	fmt.Println(userId, removeKeyRequestData.Id)
	err := usersHandler.resourcesConnection.Users.Update.RemoveKey(userId, removeKeyRequestData.Id)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(writer, err)
		return
	}
	fmt.Fprint(writer, "ok")
}

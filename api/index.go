package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"google.golang.org/api/sheets/v4"
	"searchspring.com/orgchart/model"
)

var router *mux.Router
var client = &http.Client{}

// Handler - check routing and call correct methods
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path)
	if router == nil {
		r, err := CreateRouter()
		if err != nil {
			WriteError(w, 500, err.Error())
			return
		}
		router = r
	}
	router.ServeHTTP(w, r)
}

// CreateRouter public so we can test it.
func CreateRouter() (*mux.Router, error) {
	router := mux.NewRouter()
	apiSubRouter := router.PathPrefix("/api").Subrouter()
	apiSubRouter.HandleFunc("/preview", wrapSignedInUserCheck(Preview)).Methods(http.MethodPost)
	return router, nil
}

func Preview(w http.ResponseWriter, r *http.Request, u *model.User) {
	onboard := &model.Onboard{}
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteError(w, 500, "could not read body: "+err.Error())
		return
	}
	if err := json.Unmarshal(reqBody, onboard); err != nil {
		WriteError(w, 500, "could not unmarshal body: "+err.Error())
		return
	}

	systems, err := getSystemsForRole(u, onboard.Role)
	if err != nil {
		WriteError(w, 500, "could not load sheet data: "+err.Error())
		return
	}
	onboard.Systems = systems
	outBytes, err := json.Marshal(onboard)
	if err != nil {
		WriteError(w, 500, "could not marshal results: "+err.Error())
		return
	}
	w.Write(outBytes)
}

func getSystemsForRole(user *model.User, role string) ([]*model.Task, error) {
	req, err := http.NewRequest("GET", "https://content-sheets.googleapis.com/v4/spreadsheets/11-cVOW6txhd0OsW-eeKKIyDp_Y0Jo3a4loslKoTVOXE/values/A1%3AZZ900?majorDimension=COLUMNS", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("authorization", "Bearer "+user.AccessToken)
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	if response.StatusCode != 200 {
		log.Println("error from google sheets server", string(body))
		return nil, fmt.Errorf("failed to find user for access token")
	}
	sheetsValue := &sheets.ValueRange{}
	if err := json.Unmarshal(body, sheetsValue); err != nil {
		return nil, err
	}
	systems := sheetsValue.Values[0]
	tasks := []*model.Task{}
	responsible := sheetsValue.Values[0]

	for _, column := range sheetsValue.Values {
		if len(column) > 0 && strings.ToLower(fmt.Sprintf("%s", column[0])) == "admin 1" {
			responsible = column
		}
	}
	for _, column := range sheetsValue.Values {
		if len(column) > 0 && column[0] == role {
			for i, row := range column {
				if strings.ToLower(strings.TrimSpace(fmt.Sprintf("%s", row))) == "x" {
					system := fmt.Sprintf("%s", systems[i])
					assigneeEmail := fmt.Sprintf("%s", responsible[i])
					tasks = append(tasks, &model.Task{
						Name:          system,
						AssigneeEmail: assigneeEmail,
					})

				}
			}
		}
	}

	return tasks, nil
}
func lookupAssignee(system string, tasks *model.Onboard, assignee string) string {
	// TODO implement
	if strings.ToLower(assignee) == "manager" {
		return tasks.ManagerEmail
	}
	return "tbd@example.com"
}

func wrapSignedInUserCheck(apiRequest func(w http.ResponseWriter, r *http.Request, u *model.User)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if u, err := checkUserLoggedIn(r); err != nil {
			WriteError(w, 403, err.Error())
		} else {

			apiRequest(w, r, u)
		}
	}
}

func checkUserLoggedIn(r *http.Request) (*model.User, error) {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization == "" {
		return nil, fmt.Errorf("authorization failed - no authorization header")
	}
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + authorization)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	if response.StatusCode != 200 {
		log.Println("error from google auth server", string(body))
		return nil, fmt.Errorf("failed to find user for access token")
	}
	user := &model.User{}
	err = json.Unmarshal(body, user)
	if err != nil {
		return nil, err
	}
	user.AccessToken = authorization
	return user, nil
}

// WriteError send a response with a status code.
func WriteError(w http.ResponseWriter, code int, e string) {
	w.WriteHeader(code)
	values := map[string]string{
		"error": e,
	}
	valuesBytes, _ := json.Marshal(values)
	_, err := w.Write(valuesBytes)
	if err != nil {
		log.Println(err.Error())
	}
}

package controllers

import (
	web "babou/lib/web"
	fmt "fmt"

	bcrypt "code.google.com/p/go.crypto/bcrypt"
	rand "crypto/rand"

	models "babou/app/models"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

type LoginController struct {
	actionMap map[string]web.Action
}

// Registers actions for the HomeController and returns it.
// Note: State in the returned controller object is global to
// all requests the controller processes.
func NewLoginController() *LoginController {
	lc := &LoginController{}
	lc.actionMap = make(map[string]web.Action)

	//add your actions here.
	lc.actionMap["index"] = lc.Index
	lc.actionMap["create"] = lc.Create
	lc.actionMap["new"] = lc.New

	return lc
}

func (lc *LoginController) HandleRequest(action string,
	params map[string]string) *web.Result {

	if lc.actionMap[action] != nil {
		return lc.actionMap[action](params)
	} else {
		return &web.Result{Status: 404, Body: []byte("Not found")}
	}
}

func (lc *LoginController) Index(params map[string]string) *web.Result {
	output := &web.Result{}

	output.Status = 200
	outData := &web.ViewData{Context: &struct{}{}}

	output.Body = []byte(web.RenderIn("public", "login", "index", outData))

	return output
}

func (lc *LoginController) New(params map[string]string) *web.Result {
	output := &web.Result{}

	output.Status = 200
	outData := &web.ViewData{Context: &struct{}{}}

	output.Body = []byte(web.RenderIn("public", "login", "new", outData))

	return output
}

func (lc *LoginController) Create(params map[string]string) *web.Result {
	//64-char salt
	saltLength := 64
	passwordSalt := make([]byte, saltLength)
	password := make([]byte, 0)

	n, err := rand.Read(passwordSalt)
	if n != len(passwordSalt) || err != nil {
		return &web.Result{Status: 500, Body: []byte(fmt.Sprintf("error generating random salt"))}
	}

	password = append(passwordSalt, []byte(params["password"])...)
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
	if err != nil {
		return &web.Result{Status: 500, Body: []byte(fmt.Sprintf("error encrypting password"))}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, password)
	if err != nil {
		return &web.Result{Status: 500, Body: []byte(fmt.Sprintf("error comparing password %s", err))}
	}

	username := params["username"]

	err = models.NewUser(username, hashedPassword, passwordSalt)
	if err != nil {
		fmt.Printf("error creating user: %s", err.Error())
	}

	redirectPath := &web.RedirectPath{
		NamedRoute: "loginIndex", //redirect to login page.
	}

	fmt.Printf("\n issuing redirect \n")
	return &web.Result{Status: 302, Body: nil, Redirect: redirectPath}

}

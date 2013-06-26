package controllers

import (
	web "babou/lib/web"
	fmt "fmt"

	bcrypt "code.google.com/p/go.crypto/bcrypt"
	rand "crypto/rand"

	models "babou/app/models"

	errors "errors"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

type LoginController struct {
	safeInstance bool //`true` if this instance can service HTTP requests, false otherwise.
	context      web.Context

	actionMap map[string]web.Action
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

// Registers actions for the HomeController and returns it.
func NewLoginController() *LoginController {
	lc := &LoginController{}
	lc.safeInstance = false

	return lc
}

// Implementations of DevController and Route

// Sets the login controller's context which includes POST/GET vars.
func (lc *LoginController) SetContext(context web.Context) error {
	if lc.safeInstance {
		lc.context = context
		return nil
	}

	return errors.New("This instance of LoginController cannot service requests.")
}

// Dispatches routes through this controller's actionMap and returns a result.
func (lc *LoginController) HandleRequest(action string) *web.Result {
	if !lc.safeInstance {
		return &web.Result{Status: 500, Body: []byte("Server could not route your request.")}
	}

	if lc.actionMap[action] != nil {
		return lc.actionMap[action](lc.context.GetParams())
	} else {
		return &web.Result{Status: 404, Body: []byte("Not found")}
	}
}

// Prepares a public-facing instance of this route that should be used for a single request.
func (lc *LoginController) Process(action string, context web.Context) (web.Controller, error) {
	//default route processor.
	return process(lc, action, context)
}

func (lc *LoginController) NewInstance() web.Controller {
	newLc := &LoginController{safeInstance: true, actionMap: make(map[string]web.Action)}

	//add your actions here.
	newLc.actionMap["index"] = newLc.Index
	newLc.actionMap["create"] = newLc.Create
	newLc.actionMap["new"] = newLc.New

	return newLc
}

func (lc *LoginController) IsSafeInstance() bool {
	return lc.safeInstance
}

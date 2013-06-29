package controllers

import (
	errors "errors"
	fmt "fmt"

	bcrypt "code.google.com/p/go.crypto/bcrypt"
	rand "crypto/rand"

	filters "github.com/drbawb/babou/app/filters"
	models "github.com/drbawb/babou/app/models"

	web "github.com/drbawb/babou/lib/web"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

type LoginController struct {
	safeInstance bool //`true` if this instance can service HTTP requests, false otherwise.

	context *filters.DevContext
	session *filters.SessionContext
	flash   *filters.FlashContext

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
	outData := &web.ViewData{Context: &struct{}{}} // render the registration form.

	output.Body = []byte(web.RenderWith("public", "login", "new", outData, lc.flash))

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

	// redirect to login#New() w/ flash message saying passwords don't match.
	if params["password"] != params["confirm-password"] {
		fmt.Printf("redirecting to new page; password mismatch")
		lc.flash.AddFlash("The password and confirmation do not match. Please double-check your supplied passwords.")
		redirectPath := &web.RedirectPath{
			NamedRoute: "loginNew", //redirect to login page.
		}

		return &web.Result{Status: 302, Body: nil, Redirect: redirectPath}
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

func (lc *LoginController) SetFlashContext(fc *filters.FlashContext) error {
	if fc == nil || !lc.safeInstance {
		return errors.New("Login controller or flash context not ready for request.")
	}

	lc.flash = fc

	return nil
}

func (lc *LoginController) SetSessionContext(sc *filters.SessionContext) {
	lc.session = sc
}

// Sets the login controller's context which includes POST/GET vars.
func (lc *LoginController) SetContext(context *filters.DevContext) error {
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
func (lc *LoginController) Process(action string) (web.Controller, error) {
	//default route processor.
	return process(lc, action)
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

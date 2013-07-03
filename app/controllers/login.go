package controllers

import (
	errors "errors"
	fmt "fmt"

	filters "github.com/drbawb/babou/app/filters"
	models "github.com/drbawb/babou/app/models"

	web "github.com/drbawb/babou/lib/web"
)

// Implements babou/app.Controller interface.
// Maps an action to results or returns 404 otherwise.

const (
	ACCT_CREATION_ERROR = `There was an unexpected error while creating your account. Please try again later or
	contact our administrative staff.`
)

type LoginController struct {
	safeInstance bool //`true` if this instance can service HTTP requests, false otherwise.

	context *filters.DevContext
	session *filters.SessionContext
	flash   *filters.FlashContext
	auth    *filters.AuthContext

	actionMap map[string]web.Action
}

func (lc *LoginController) Index(params map[string]string) *web.Result {
	output := &web.Result{}

	output.Status = 200
	outData := &web.ViewData{Context: &struct{}{}}

	output.Body = []byte(web.RenderWith("public", "login", "index", outData, lc.flash))

	return output
}

func (lc *LoginController) Session(params map[string]string) *web.Result {
	output := &web.Result{}
	// otherwise redirect those bitches to loginPage with an error.
	redirectPath := &web.RedirectPath{
		NamedRoute: "homeIndex", //redirect to login page.
	}

	// check credentials and get user.
	user := &models.User{}
	err := user.SelectUsername(params["username"])
	if err != nil {
		lc.flash.AddFlash(fmt.Sprintf("Error logging you in: %s", err.Error()))
		output.Status = 302

		redirectPath.NamedRoute = "loginIndex"
		output.Redirect = redirectPath
		// redirect them back to the login page w/ an error.

		return output
	}

	err = user.CheckHash(params["password"])
	if err != nil {
		lc.flash.AddFlash(fmt.Sprintf("Error logging you in: %s", err.Error()))
		output.Status = 302
		output.Redirect = redirectPath
		// redirect them to the home-page which should now render as a news page.
		return output
	}

	session, _ := lc.session.GetSession()

	session.Values["user_id"] = user.UserId

	output.Status = 302
	output.Redirect = redirectPath

	lc.auth.WriteSessionFor(user)

	return output
}

func (lc *LoginController) Logout(params map[string]string) *web.Result {
	output := &web.Result{}
	// otherwise redirect those bitches to loginPage with an error.
	redirectPath := &web.RedirectPath{
		NamedRoute: "homeIndex", //redirect to login page.
	}

	// check credentials and get user.
	//user := &models.User{}
	//err := user.SelectUsername(params["username"])
	err := lc.auth.DeleteCurrentSession()
	if err != nil {
		lc.flash.AddFlash(fmt.Sprintf("Error logging you out: %s", err.Error()))
		output.Status = 302

		output.Redirect = redirectPath
		// redirect them back to the login page w/ an error.

		return output
	}

	output.Status = 302
	output.Redirect = redirectPath

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

	username, password := params["username"], params["password"]
	// redirect to login#New() w/ flash message saying passwords don't match.
	if params["password"] != params["confirm-password"] {
		fmt.Printf("redirecting to new page; password mismatch")
		lc.flash.AddFlash("the password and confirmation you entered do not match. Please double-check your supplied passwords.")
		redirectPath := &web.RedirectPath{
			NamedRoute: "loginNew", //redirect to login page.
		}

		return &web.Result{Status: 302, Body: nil, Redirect: redirectPath}
	}

	status, err := models.NewUser(username, password)
	if err != nil {
		return &web.Result{Status: 500, Body: []byte(ACCT_CREATION_ERROR)}
	}

	redirectPath := &web.RedirectPath{
		NamedRoute: "loginIndex", //redirect to login page.
	}

	// Redirect back to registration page if there was an error creating account.
	if status == models.USERNAME_TAKEN {
		lc.flash.AddFlash("The username you chose was already taken")
		redirectPath.NamedRoute = "loginNew"
	} else if status != 0 {
		lc.flash.AddFlash("There was an error validating your new user account; please try again or contact our administrative staff.")
		redirectPath.NamedRoute = "loginNew"
	} else {
		//TODO: show message if account activation is required.
		lc.flash.AddFlash("Your account was created sucesfully. You may now login.")
	}

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

func (lc *LoginController) SetSessionContext(sc *filters.SessionContext) error {
	lc.session = sc
	return nil
}

// Sets the login controller's context which includes POST/GET vars.
func (lc *LoginController) SetContext(context *filters.DevContext) error {
	if lc.safeInstance {
		lc.context = context
		return nil
	}

	return errors.New("This instance of LoginController cannot service requests.")
}

func (lc *LoginController) SetAuthContext(context *filters.AuthContext) error {
	if lc.safeInstance {
		lc.auth = context
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

// Tests that the current chain is sufficient for this route.
func (lc *LoginController) TestContext(chain []web.ChainableContext) error {
	outFlag := false
	for i := 0; i < len(chain); i++ {
		_, ok := chain[i].(filters.AuthChainLink)
		if ok {
			outFlag = true
			break
		}
	}

	if err := testContext(chain); err != nil {
		return errors.New("Default chain missing from login route")
	}

	if !outFlag {
		return errors.New("Auth chain missing from login route.")
	}

	return nil
}

func (lc *LoginController) NewInstance() web.Controller {
	newLc := &LoginController{safeInstance: true, actionMap: make(map[string]web.Action)}

	//add your actions here.
	newLc.actionMap["index"] = newLc.Index
	//registration
	newLc.actionMap["create"] = newLc.Create
	newLc.actionMap["new"] = newLc.New
	//session
	newLc.actionMap["session"] = newLc.Session
	newLc.actionMap["logout"] = newLc.Logout
	return newLc
}

func (lc *LoginController) IsSafeInstance() bool {
	return lc.safeInstance
}

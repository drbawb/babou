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

	*App
	auth *filters.AuthContext

	actionMap map[string]web.Action
}

// Dispatch() makes this controller implement a dispatcher.
//
// The dispatcher instantiates a request-facing instance of this
// controller and returns its action or nil if the action is not found.
func (lc *LoginController) Dispatch(action, accept string) (web.Controller, web.Action) {
	newLc := &LoginController{
		safeInstance: true,
		actionMap:    make(map[string]web.Action),
		App:          &App{},
	}

	//login page.
	newLc.actionMap["index"] = newLc.Index

	//registration
	newLc.actionMap["create"] = newLc.Create
	newLc.actionMap["new"] = newLc.New

	//session
	newLc.actionMap["session"] = newLc.Session
	newLc.actionMap["logout"] = newLc.Logout

	return newLc, newLc.actionMap[action]
}

// Displays the login page.
func (lc *LoginController) Index() *web.Result {
	output := &web.Result{}

	output.Status = 200
	outData := &web.ViewData{Context: &struct{}{}}

	output.Body = []byte(web.RenderWith("bootstrap", "login", "index", outData, lc.Flash))

	return output
}

// Creates a new session after a user attempts a login.
func (lc *LoginController) Session() *web.Result {
	output := &web.Result{}
	// otherwise redirect those bitches to loginPage with an error.
	redirectPath := &web.RedirectPath{
		NamedRoute: "homeIndex", //redirect to login page.
	}

	// check credentials and get user.
	user := &models.User{}
	err := user.SelectUsername(lc.Dev.Params.All["username"])
	if err != nil {
		lc.Flash.AddFlash(fmt.Sprintf("Error logging you in: %s", err.Error()))
		output.Status = 302

		redirectPath.NamedRoute = "loginIndex"
		output.Redirect = redirectPath
		// redirect them back to the login page w/ an error.

		return output
	}

	err = user.CheckHash(lc.Dev.Params.All["password"])
	if err != nil {
		fmt.Printf("error logging you in.")
		lc.Flash.AddFlash(fmt.Sprintf("Error logging you in: %s", err.Error()))
		output.Status = 302

		redirectPath.NamedRoute = "loginIndex"
		output.Redirect = redirectPath
		// redirect them to the home-page which should now render as a news page.
		return output
	}

	session, _ := lc.App.Session.GetSession()

	session.Values["user_id"] = user.UserId

	output.Status = 302
	output.Redirect = redirectPath

	lc.auth.WriteSessionFor(user)

	return output
}

// Removes a users session from the database. Removes their session cookie.
// Then redirects them to the homepage. (Which should now show the welcome banner.)
func (lc *LoginController) Logout() *web.Result {
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
		//lib.Println("error logging user out: %v \n", err.Error())
		lc.Flash.AddFlash(fmt.Sprintf("Error logging you out: %s", err.Error()))
		output.Status = 302

		output.Redirect = redirectPath
		// redirect them back to the login page w/ an error.

		return output
	}

	output.Status = 302
	output.Redirect = redirectPath

	return output
}

// Displays the registration form.
func (lc *LoginController) New() *web.Result {
	output := &web.Result{}

	output.Status = 200
	outData := &web.ViewData{Context: &struct{}{}} // render the registration form.

	output.Body = []byte(web.RenderWith("bootstrap", "login", "new", outData, lc.Flash))

	return output
}

// Handles the results from the registration form submission.
func (lc *LoginController) Create() *web.Result {

	username, password := lc.Dev.Params.All["username"], lc.Dev.Params.All["password"]
	// redirect to login#New() w/ flash message saying passwords don't match.
	if lc.Dev.Params.All["password"] != lc.Dev.Params.All["confirm-password"] {
		fmt.Printf("redirecting to new page; password mismatch")
		lc.Flash.AddFlash("the password and confirmation you entered do not match. Please double-check your supplied passwords.")
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
		lc.Flash.AddFlash("The username you chose was already taken")
		redirectPath.NamedRoute = "loginNew"
	} else if status != 0 {
		lc.Flash.AddFlash("There was an error validating your new user account; please try again or contact our administrative staff.")
		redirectPath.NamedRoute = "loginNew"
	} else {
		//TODO: show message if account activation is required.
		lc.Flash.AddFlash("Your account was created sucesfully. You may now login.")
	}

	return &web.Result{Status: 302, Body: nil, Redirect: redirectPath}
}

// Returns a LoginController instance that is not safe across requests.
// This is useful for routing as well as context-testing.
func NewLoginController() *LoginController {
	lc := &LoginController{}
	lc.safeInstance = false

	return lc
}

// Sets up contexts.
func (lc *LoginController) SetAuthContext(context *filters.AuthContext) error {
	if lc.safeInstance {
		lc.auth = context
		return nil
	}

	return errors.New("This instance of LoginController cannot service requests.")
}

// Tests that the current chain is sufficient for this route.
// This route requires the default chain as well as the authorization context.
func (lc *LoginController) TestContext(chain []web.ChainableContext) error {
	lc.App.TestContext(chain)

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

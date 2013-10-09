package filters

import (
	"errors"
	"fmt"
	"net/http"

	models "github.com/drbawb/babou/app/models"
	web "github.com/drbawb/babou/lib/web"
)

type AuthChainLink interface {
	web.ChainableContext
	WriteSessionFor(*models.User) error
	DeleteCurrentSession() error
}

// An authorizable controller must be willing to accept an AuthContext.
//
// This AuthContext will be tied to a session for a single request.
type AuthorizableController interface {
	web.Controller
	SetAuthContext(*AuthContext) error
}

// An implementation of SessionContext that uses it to provide helper methods for authorizing a user.
type AuthContext struct {
	isInit   bool
	required bool

	request  *http.Request
	response http.ResponseWriter

	session SessionChainLink
}

// Returns an uninitialized AuthContext suitable for use in a context chain
//
// `required: true` will require that a current sesion meet some criteria
// otherwise this chain will stop the response during the AFTER-ATTACH resolution
/// phase
func AuthChain(required bool) *AuthContext {
	context := &AuthContext{isInit: false, required: required}

	return context
}

func (ac *AuthContext) DeleteCurrentSession() error {
	session, _ := ac.session.GetSession()

	var userId int
	if session.Values["user_id"] == nil {
		return errors.New("You are not currently logged in.")
	} else {
		userId = session.Values["user_id"].(int)
	}

	user := &models.User{}
	if err := user.SelectId(userId); err != nil {
		return err
	}

	if err := ac.deleteSessionFor(user); err != nil {
		return err
	}

	session.Values["user_id"] = nil
	return nil
}

// Delete a session when the user logs out.
// This will fail all future permission checks for that user.
func (ac *AuthContext) deleteSessionFor(user *models.User) error {
	if user == nil || ac.isInit == false {
		return errors.New("This auth-context is not ready to delete a user's session.")
	}

	userSession := &models.Session{}
	err := userSession.DeleteFor(user)
	if err != nil {
		return err
	}

	return nil
}

// Only doing this b/c AC has access to request/response
// Some way I can make this private to `login` controller?
func (ac *AuthContext) WriteSessionFor(user *models.User) error {
	if user == nil || ac.isInit == false {
		return errors.New("This auth-context is not ready to write a user session.")
	}

	fmt.Printf("writing a session for user: %s w/ IP: %s \n", user.Username, ac.request.RemoteAddr)
	userSession := &models.Session{}
	err := userSession.WriteFor(user, ac.request.RemoteAddr)
	if err != nil {
		fmt.Printf("[auth-context] error saving user session: %s \n", err.Error())
		return err
	}

	return nil
	//fmt.Printf("Writing a sessino for user: %s w/ remote IP: %s", user.Username, ac.request.RemoteAddr
}

// Returns the currently authenticated user
func (ac *AuthContext) CurrentUser() (*models.User, error) {
	session, _ := ac.session.GetSession()
	userId := session.Values["user_id"]

	if userId == nil {
		return nil, errors.New("No current user.")
	}

	user := &models.User{}
	if err := user.SelectId(userId.(int)); err != nil {
		return nil, err
	}

	return user, nil
}

// Checks if a user can perform a specified action:
func (ac *AuthContext) Can(permission string) bool {
	// if the user has a session: it will return true for every check.
	// otherwise it will return false.
	session, err := ac.session.GetSession()
	if err != nil {
		fmt.Printf("Error authorizing: %s", err.Error())
		return false
	}

	if session.Values["user_id"] != nil {
		return true
	} else {
		return false
	}

	// hah, i dont care riiight now.
}

// Requires authentication based on request/response
// This requires authentication without involving your controller.
//
// Useful for protecting groups of routes.
//
// The AuthContext will only call this if it is initialized with
// `AuthContext.required = true`
func (ac *AuthContext) AfterAttach(w http.ResponseWriter, r *http.Request) error {
	if !ac.required {
		return nil
	}

	user, err := ac.CurrentUser()
	if err != nil || user == nil {
		return errors.New("YOU MUST BE LOGGED IN TO VIEW THIS PAGE.")
	} else if !user.IsAdmin {
		return errors.New("YOU ARE NOT AUTHORIZED TO VIEW THIS PAGE.")
	} else {
		return nil
	}
}

// No-op
func (ac *AuthContext) CloseContext() {}

// This context requires a chain with a SessionChainLink as well as an AuthorizableController route.
//
// This method can be used to ensure that those dependencies are satisfied at runtime.
func (ac *AuthContext) TestContext(route web.Controller, chain []web.ChainableContext) error {
	//requires AuthorizableController and SessionChain
	hasSession := false

	_, ok := route.(AuthorizableController)
	if !ok {
		return errors.New(fmt.Sprintf("The route :: %T :: does not support the AuthContext.", route))
	}

	for i := 0; i < len(chain); i++ {
		_, ok := chain[i].(SessionChainLink)
		if ok {
			hasSession = true
		}
	}

	if hasSession && ok {
		return nil
	} else {
		return errors.New(fmt.Sprintf("The route :: %T :: does not have a SessionAware context in it's context chain.", route))
	}
}

// Returns a clean instance of AuthContext that can be used safely for a single request.
func (ac *AuthContext) NewInstance() web.ChainableContext {
	newAc := &AuthContext{isInit: false, required: ac.required}

	return newAc
}

// Applies the context to an authorizable controller.
func (ac *AuthContext) ApplyContext(controller web.Controller, response http.ResponseWriter, request *http.Request, chain []web.ChainableContext) {
	ac.isInit = true

	ac.request = request
	ac.response = response

	v, ok := controller.(AuthorizableController)
	if ok {
		if err := v.SetAuthContext(ac); err != nil {
			fmt.Printf("Error setting authorization context: %s \n", err.Error())
		}
	} else {
		fmt.Printf("Tried to wrap a controller that is not AuthContext aware \n")
	}

getSession:
	for i := 0; i < len(chain); i++ {
		v, ok := chain[i].(SessionChainLink)
		if ok {
			ac.session = v

			// access session safely in here for user_id perm checks.

			break getSession
		}
	}
}

// Returns `true` if the AuthContext is properly initialized on top of a session store.
func (ac *AuthContext) isValid() bool {
	return ac.isInit
}

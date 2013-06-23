// Contains the view rendering helpers.
package web

import (
	fmt "fmt"

	mustache "github.com/drbawb/mustache"
	template "html/template"

	mux "github.com/gorilla/mux"
	url "net/url"

	errors "errors"
)

var Router *mux.Router = nil

// Layout
type ViewData struct {
	Yield   func(params []string, data string) string
	Context interface{}
}

type viewHelpers struct {
	LinkTo  func(params []string, data string) string
	UrlFor  func(params []string, data string) string
	FormFor func(params []string, data string) string
}

type postHelpers struct {
	LabelFor     func(params []string, data string) string
	TextFieldFor func(params []string, data string) string
}

func getHelpers() *viewHelpers {
	return &viewHelpers{LinkTo: LinkTo, UrlFor: UrlFor, FormFor: BuildForm}
}

func getFormHelpers() *postHelpers {
	return &postHelpers{LabelFor: LabelFor, TextFieldFor: TextFieldFor}
}

//Renders the application template; yielding to app layout if requested.
func Render(controllerName, actionName string, viewData *ViewData) string {
	layoutFile := fmt.Sprintf("app/views/%s.template", "application")
	filename := fmt.Sprintf("app/views/%s/%s.template", controllerName, actionName)

	yieldFn := func(params []string, data string) string {
		yieldOut := mustache.RenderFile(filename, viewData.Context)

		return yieldOut
	}

	viewData.Yield = yieldFn

	out := mustache.RenderFile(layoutFile, viewData)

	return out
}

func RenderIn(templateName, controllerName, actionName string, viewData *ViewData) string {
	layoutFile := fmt.Sprintf("app/views/%s.template", templateName)
	filename := fmt.Sprintf("app/views/%s/%s.template", controllerName, actionName)

	yieldFn := func(params []string, data string) string {
		yieldOut := mustache.RenderFile(filename, viewData.Context, getHelpers(), getFormHelpers())

		return yieldOut
	}

	viewData.Yield = yieldFn

	out := mustache.RenderFile(layoutFile, viewData, getHelpers())

	return out
}

// Renders a mustache template from the views/ directory
func RenderTo(templateName string, viewData *ViewData) string {
	filename := "app/views/" + templateName + ".template"
	out := mustache.RenderFile(filename, viewData.Context, getHelpers())

	return out
}

// Below here are the implementations of various view helpers

// Returns an HTML link as a string suitable for insertion into an HTML template.
// Params: {{#UrlFor [controllerName] [routeParameters]...}}
// 	params are passed to the router to create a properly formed URL.
// Data: {{#UrlFor}} [data] {{/UrlFor}}
// 	data will be escaped and used as the link's display text.
func LinkTo(params []string, data string) string {
	// Recover from bad route
	url, displayText, err := buildUrl(params, data)

	if err != nil {
		fmt.Printf("Error building URL: %s", err)
		return ""
	} else {
		return fmt.Sprintf("<a href=\"%s\">%s</a>",
			url.Path, template.HTMLEscapeString(displayText))
	}

}

// Returns a URL as a string suitable for insertion into an HTML template.
// Params: {{#UrlFor [controllerName] [routeParameters]...}}
// params are passed to the router to create a properly formed URL.
func UrlFor(params []string, data string) string {
	url, _, err := buildUrl(params, data)

	if err != nil {
		fmt.Printf("Error building URL: %s", err)
		return ""
	} else {
		return fmt.Sprintf("%s", template.HTMLEscapeString(url.Path))
	}
}

// Generates URL for a matched route.
func buildUrl(params []string, data string) (*url.URL, string, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic'd while building URL.")
			fmt.Printf("\n %s", r)

		}
	}()

	var controllerName string
	var url *url.URL
	var err error

	if len(params) > 1 {
		controllerName = params[0]
		url, err = Router.Get(controllerName).URL(params[1:]...)
		return url, data, err
	} else if len(params) == 1 {
		controllerName = params[0]
		url, err = Router.Get(controllerName).URL()
		return url, data, err
	} else {
		return nil, "", errors.New(fmt.Sprintf("No route for params: %s", params))
	}
}

func buildUrlWithVerb(controllerName string, httpMethod string) (*url.URL, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic'd while building URL.")
			fmt.Printf("\n %s", r)

		}
	}()

	baseRoute := Router.Get(controllerName)
	if baseRoute == nil {
		return nil, errors.New(fmt.Sprintf("Could not find controller: %s",
			controllerName))
	}

	baseRoute = baseRoute.Methods(httpMethod)
	if baseRoute == nil {
		return nil, errors.New(fmt.Sprintf("Controller would not respond to: %s",
			httpMethod))
	}

	//TODO: pass addt'l parameters?
	url, err := baseRoute.URL()

	return url, err
}

// Builds a form context and renders a sub-template inside of it.
func BuildForm(params []string, data string) string {
	var controllerName string
	var httpMethod string
	var formId string

	// Parse parameters
	if len(params) == 3 {
		controllerName = params[0]
		httpMethod = params[1]
		formId = params[2]
	}
	if len(params) == 2 {
		controllerName = params[0]
		formId = params[0]
		httpMethod = params[1]
	} else if len(params) == 1 {
		controllerName = params[0]
		formId = params[0]
		httpMethod = "post" //default to post for a form.
	}

	// Try to generate a route for the form results
	url, err := buildUrlWithVerb(controllerName, httpMethod)
	if err != nil {
		fmt.Printf("Unable to build a route for: %s\n debug err: %s", params, err)
		return ""
	}

	// Opening and closing form tags
	openTag := fmt.Sprintf("<form id=\"%s\" action=\"%s\" method=\"%s\">", formId, url, httpMethod)
	closeTag := fmt.Sprintf("</form>")

	// Render inner-content with Form context

	formBody := mustache.Render(data, getFormHelpers())
	return fmt.Sprintf("%s\n%s\n%s", openTag, formBody, closeTag)
}

func LabelFor(params []string, data string) string {
	var fieldId string
	var body string

	if len(params) > 0 {
		fieldId = params[0]
		body = data

		return fmt.Sprintf("<label for=\"%s\">%s</label>", fieldId, body)
	} else {
		return ""
	}
}

func TextFieldFor(params []string, data string) string {
	var fieldId string
	var fieldType string

	if len(params) > 1 {
		fieldId = params[0]
		fieldType = params[1]

		return fmt.Sprintf("<input id=\"%s\" name=\"%s\" type=\"%s\"></label>", fieldId, fieldId, fieldType)
	} else if len(params) == 1 {
		fieldId = params[0]

		return fmt.Sprintf("<input id=\"%s\" name=\"%s\" type=\"text\"></label>", fieldId, fieldId)
	} else {
		return ""
	}
}

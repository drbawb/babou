// Contains the view rendering helpers.
package web

import (
	fmt "fmt"
	mustache "github.com/drbawb/mustache"
)

// Layout
type ViewData struct {
	Yield   func(in string) string
	Context interface{}
}

//Renders the application template; yielding to app layout if requested.
func Render(controllerName, actionName string, viewData *ViewData) string {
	layoutFile := fmt.Sprintf("app/views/%s.template", "application")
	filename := fmt.Sprintf("app/views/%s/%s.template", controllerName, actionName)

	yieldFn := func(in string) string {
		yieldOut := mustache.RenderFile(filename, viewData.Context)

		return yieldOut
	}

	viewData.Yield = yieldFn

	out := mustache.RenderFile(layoutFile, viewData)

	return out
}

// Renders a mustache template from the views/ directory
func RenderTo(templateName string, viewData *ViewData) string {

	filename := "app/views/" + templateName + ".template"
	out := mustache.RenderFile(filename, viewData)

	return out
}

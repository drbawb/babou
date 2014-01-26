package web

import (
	fmt "fmt"

	mustache "github.com/drbawb/mustache"
	mux "github.com/gorilla/mux"
)

const DEBUG_RELOAD = false

// A global router or middleware implementation that will service requests
// from the HTTP server and direct them to an appropriate controller
var Router *mux.Router = nil

// A renderer maps actions to executable views.
type Renderer interface {
	RenderWith(string, string, string, ...interface{}) string
}

type MustacheRenderer struct {
	viewsRoot string
}

func NewMustacheRenderer(viewsPath string) *MustacheRenderer {
	mr := &MustacheRenderer{viewsRoot: viewsPath}

	return mr
}

// A context that will be passed to the underlying html template.
// Yield is a function that will be replaced by the renderer. It will call
// your requested template and automatically pass it the supplied `Context`
// argument.
type ViewData struct {
	Yield func(params []string, data string) string
	//Flash   func() string
	Context interface{}
}

// Helper functions which are available to views that are rendered using
// the built-in methods.
type viewHelpers struct {
	LinkTo  func(params []string, data string) string
	UrlFor  func(params []string, data string) string
	FormFor func(params []string, data string) string
}

// Helper functions which are available to views that are rendered inside
// the context of a {{#FormFor}} section.
type postHelpers struct {
	FileFor      func(params []string, data string) string
	LabelFor     func(params []string, data string) string
	TextFieldFor func(params []string, data string) string
}

var templateCache map[string]*mustache.Template = make(map[string]*mustache.Template)

// Returns a set of viewHelpers to be passed to the rendering context.
func getHelpers() *viewHelpers {
	return &viewHelpers{LinkTo: LinkTo, UrlFor: UrlFor, FormFor: BuildForm}
}

// Returns a set of form helpers to be passed to a rendering context which
// is processing an HTML form.
func getFormHelpers() *postHelpers {
	return &postHelpers{FileFor: FileFor, LabelFor: LabelFor, TextFieldFor: TextFieldFor}
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

// Renders the requested template inside a layout. This can override the
// default behavior to render inside the application layout.
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

//DEV: using some caching to hopefully cut down file i/o

func RenderWith(templateName, controllerName, actionName string, filterHelpers ...interface{}) string {
	return renderWith("app/views", templateName, controllerName, actionName, filterHelpers...)
}

func (mu *MustacheRenderer) RenderWith(
	templateName,
	controllerName,
	actionName string,
	filterHelpers ...interface{}) string {

	return renderWith(mu.viewsRoot, templateName, controllerName, actionName, filterHelpers...)
}

func renderWith(viewsRoot, templateName, controllerName, actionName string, filterHelpers ...interface{}) string {
	layoutFile := fmt.Sprintf("%s/%s.template", viewsRoot, templateName)
	filename := fmt.Sprintf("%s/%s/%s.template", viewsRoot, controllerName, actionName)

	//placeholder for dev.

	expandedFilterHelpers := make([]interface{}, 0)

	yieldFn := func(params []string, data string) string {
		if templateCache[filename] == nil {
			var err error

			templateCache[filename], err = mustache.ParseFile(filename)
			if err != nil {
				return "Error preparing template."
			}
		}

		yieldOut := templateCache[filename].Render(expandedFilterHelpers...)

		return yieldOut
	}

	viewData := &struct {
		Yield func([]string, string) string
	}{
		Yield: yieldFn,
	}

	expandedFilterHelpers = append(expandedFilterHelpers, viewData, getHelpers(), getFormHelpers())
	for i := 0; i < len(filterHelpers); i++ {
		v, ok := filterHelpers[i].(ViewableContext)
		if ok {
			expandedFilterHelpers = append(expandedFilterHelpers, v.GetViewHelpers()...)
		} else {
			expandedFilterHelpers = append(expandedFilterHelpers, filterHelpers[i])
		}
	}

	if templateCache[layoutFile] == nil {
		var err error
		templateCache[layoutFile], err = mustache.ParseFile(layoutFile)

		if err != nil {
			return "Error parsing template file"
		}
	}
	out := templateCache[layoutFile].Render(expandedFilterHelpers...)

	return out
}

// Renders a mustache template from the views/ directory
func RenderTo(templateName string, viewData *ViewData) string {
	filename := "app/views/" + templateName + ".template"
	out := mustache.RenderFile(filename, viewData.Context, getHelpers())

	return out
}

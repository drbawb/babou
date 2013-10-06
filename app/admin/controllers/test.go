package admin

import (
	"github.com/drbawb/babou/app/controllers"
	"github.com/drbawb/babou/app/models"

	"github.com/drbawb/babou/lib/web"
)

type AdminTestController struct {
	*controllers.App
}

func (ap *AdminTestController) Dispatch(action string) (web.Controller, web.Action) {
	if ap.App == nil {
		ap.App = &controllers.App{}
	}

	res := &web.Result{}
	res.Status = 200

	users := &models.User{}
	//err := users.SelectList()
	/*if err != nil {
		res.Body = []byte(err.Error())

	}*/

	res.Body = []byte("free ban! one of (")

	fn := func() *web.Result {
		return res
	}

	return ap, fn
}

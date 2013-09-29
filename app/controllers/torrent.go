package controllers

import (
	filters "github.com/drbawb/babou/app/filters"
	models "github.com/drbawb/babou/app/models"

	libTorrent "github.com/drbawb/babou/lib/torrent"
	web "github.com/drbawb/babou/lib/web"

	"github.com/drbawb/babou/bridge"

	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TorrentController struct {
	safeInstance bool //`true` if this instance can service HTTP requests, false otherwise.

	context *filters.DevContext
	session *filters.SessionContext
	flash   *filters.FlashContext
	auth    *filters.AuthContext
	events  *filters.EventContext

	actionMap map[string]web.Action
}

func (tc *TorrentController) Index(params map[string]string) *web.Result {
	// Private page.
	redirect, user := tc.RedirectOnAuthFail()
	if user == nil {
		return redirect
	}

	output := &web.Result{Status: 200}
	outData := &struct {
		Username    string
		TorrentList []*models.Torrent
	}{
		Username: user.Username,
	}

	allTorrents := &models.Torrent{}
	torrentList, err := allTorrents.SelectSummaryPage()
	if err != nil {
		output.Body = []byte(err.Error())
		return output
	}

	outData.TorrentList = torrentList

	output.Body = []byte(web.RenderWith("bootstrap", "torrent", "index", outData, tc.flash))

	return output
}

// Displays a form where a user can upload a new torrent.
func (tc *TorrentController) New(params map[string]string) *web.Result {
	//TODO: permission check; for now any authenticated user can add torrents.
	redirect, user := tc.RedirectOnAuthFail()
	if user == nil {
		return redirect
	}

	output := &web.Result{Status: 200}
	outData := &struct {
		Username    string
		AnnounceURL string
	}{
		Username:    user.Username,
		AnnounceURL: user.AnnounceURL(),
	}

	// Display new torrent form.
	output.Body = []byte(web.RenderWith("bootstrap", "torrent", "new", outData, tc.flash))

	return output
}

func (tc *TorrentController) Create(params map[string]string) *web.Result {
	//TODO: permission check; for now any authenticated user can add torrents.
	redirect, user := tc.RedirectOnAuthFail()
	if user == nil {
		return redirect
	}

	outData := &struct {
		Username string
	}{
		Username: user.Username,
	}

	formFiles := tc.context.GetParams().Files
	if formFiles["metainfo"] == nil {
		tc.flash.AddFlash("File upload appears to be missing.")
	} else if len(formFiles["metainfo"]) <= 0 || len(formFiles["metainfo"]) > 1 {
		tc.flash.AddFlash("You are only allowed to upload one torrent at a time.")
	} else {
		file := formFiles["metainfo"][0]
		if !strings.HasSuffix(file.Filename, ".torrent") {
			tc.flash.AddFlash("File needs to end w/ .torrent!")
			return tc.RedirectOnUploadFail()
		}

		fmt.Printf("reading torrent: \n")
		torrentFile, err := file.Open()
		if err != nil {
			tc.flash.AddFlash("Error reading your torrent; please try your upload again")
			return tc.RedirectOnUploadFail()
		}

		torrent := libTorrent.ReadFile(torrentFile)
		torrentRecord := &models.Torrent{}

		if err := torrentRecord.Populate(torrent.Info); err != nil {
			tc.flash.AddFlash("Error reading your torrent file.")
			return tc.RedirectOnUploadFail()
		}

		attributes := &models.Attribute{}
		attributes.AlbumName = params["albumName"]
		attributes.ReleaseYear = time.Now()
		attributes.ArtistName = strings.Split(params["artistName"], ",")
		fmt.Printf("num artists: %d \n", len(attributes.ArtistName))

		torrentRecord.SetAttributes(attributes)

		if err := torrentRecord.Write(); err != nil {
			tc.flash.AddFlash(fmt.Sprintf("Error saving your torrent. Please contact a staff member: %s", err.Error()))
			return tc.RedirectOnUploadFail()
		}

		tc.flash.AddFlash(fmt.Sprintf(`Your torrents URL is: http://tracker.fatalsyntax.com/torrents/download/%d -- 
			please save this because babou cannot find things right now.`, torrentRecord.ID))
	}

	output := &web.Result{
		Status: 200,
	}

	output.Body = []byte(web.RenderWith("application", "torrent", "index", outData, tc.flash))
	return output
}

func (tc *TorrentController) Download(params map[string]string) *web.Result {
	redirect, user := tc.RedirectOnAuthFail()
	if user == nil {
		return redirect
	}

	output := &web.Result{
		Status: 200,
	}

	record := &models.Torrent{}

	var torrentId int64
	torrentId, err := strconv.ParseInt(params["torrentId"], 10, 32)

	if err != nil {
		output.Body = []byte("invalid torrent id.")
		return output
	}

	record.SelectId(int(torrentId))
	outFile, err := record.WriteFile(user.Secret, user.SecretHash)
	if err != nil {
		result := &web.Result{}
		tc.flash.AddFlash("Could not find the torrent with the specified ID")
		result.Redirect = &web.RedirectPath{}
		result.Redirect.NamedRoute = "torrentIndex"

		result.Status = 302

		return result
	}

	output.IsFile = true
	output.Filename = "test.torrent"
	output.Body = outFile

	//TODO: Test attributes.
	attributes, err := record.Attributes()
	if attributes != nil {
		fmt.Printf("attributes: %v \n", attributes)
	}

	return output
}

func (tc *TorrentController) Delete(params map[string]string) *web.Result {
	redirect, user := tc.RedirectOnAuthFail()
	if user == nil {
		return redirect
	}

	output := &web.Result{
		Status: 200,
	}

	fmt.Printf("sending pretend delete event to tracker(s) \n")
	payload := &bridge.DeleteTorrentMessage{}
	payload.InfoHash = "0xDEADBEEF"
	payload.Reason = "No logs on FLAC torrent."
	msg := &bridge.Message{Type: bridge.DELETE_TORRENT, Payload: payload}

	tc.events.SendMessage(msg)

	return output
}

// Tests if the user is logged in.
// If not: returns a web.Result that would redirect them to the homepage.
func (tc *TorrentController) RedirectOnAuthFail() (*web.Result, *models.User) {
	if user, err := tc.auth.CurrentUser(); err != nil {
		result := &web.Result{}

		result.Redirect = &web.RedirectPath{}
		result.Redirect.NamedRoute = "homeIndex"

		result.Status = 302
		return result, nil
	} else {
		return nil, user
	}
}

func (tc *TorrentController) RedirectOnUploadFail() *web.Result {
	result := &web.Result{}

	result.Redirect = &web.RedirectPath{}
	result.Redirect.NamedRoute = "torrentNew"

	result.Status = 302

	return result
}

// Returns a routable instance of TorrentController
func NewTorrentController() *TorrentController {
	tc := &TorrentController{}
	tc.safeInstance = false

	return tc
}

// Implementations of DevController and Route

func (tc *TorrentController) SetFlashContext(fc *filters.FlashContext) error {
	if fc == nil || !tc.safeInstance {
		return errors.New("Torrent controller or flash context not ready for request.")
	}

	tc.flash = fc

	return nil
}

func (tc *TorrentController) SetSessionContext(sc *filters.SessionContext) error {
	tc.session = sc

	return nil
}

// Sets the login controller's context which includes POST/GET vars.
func (tc *TorrentController) SetContext(context *filters.DevContext) error {
	if tc.safeInstance {
		tc.context = context
		return nil
	}

	return errors.New("This instance of TorrentController cannot service requests.")
}

func (tc *TorrentController) SetAuthContext(context *filters.AuthContext) error {
	if tc.safeInstance {
		tc.auth = context
		return nil
	}

	return errors.New("This instance of TorrentController cannot service requests.")
}

func (tc *TorrentController) SetEventContext(context *filters.EventContext) error {
	if tc.safeInstance {
		tc.events = context
		return nil
	}

	return errors.New("This instance of TorrentController cannot service requests.")
}

// Dispatches routes through this controller's actionMap and returns a result.
func (tc *TorrentController) HandleRequest(action string) *web.Result {
	if !tc.safeInstance {
		return &web.Result{Status: 500, Body: []byte("Server could not route your request.")}
	}

	if tc.actionMap[action] != nil {
		return tc.actionMap[action](tc.context.GetParams().All)
	} else {
		return &web.Result{Status: 404, Body: []byte("Not found")}
	}
}

// Prepares a public-facing instance of this route that should be used for a single request.
func (tc *TorrentController) Process(action string) (web.Controller, error) {
	//default route processor.
	return process(tc, action)
}

// Tests that the current chain is sufficient for this route.
func (tc *TorrentController) TestContext(chain []web.ChainableContext) error {
	outFlag := false
	eventFlag := false
	for i := 0; i < len(chain); i++ {
		if outFlag && eventFlag {
			break
		}

		_, ok := chain[i].(filters.AuthChainLink)
		if ok {
			outFlag = true
			continue
		}
		_, ok = chain[i].(filters.EventChainLink)
		if ok {
			eventFlag = true
			continue
		}
	}

	if err := testContext(chain); err != nil {
		return errors.New("Default chain missing from torrent route")
	}

	if !outFlag {
		return errors.New("Auth chain missing from torrent route.")
	}

	if !eventFlag {
		return errors.New("Event chian missing from torrent route")
	}

	return nil
}

func (tc *TorrentController) NewInstance() web.Controller {
	newTc := &TorrentController{safeInstance: true, actionMap: make(map[string]web.Action)}

	//add your actions here.
	newTc.actionMap["index"] = newTc.Index

	newTc.actionMap["new"] = newTc.New
	newTc.actionMap["create"] = newTc.Create

	newTc.actionMap["download"] = newTc.Download

	newTc.actionMap["delete"] = newTc.Delete

	return newTc
}

func (tc *TorrentController) IsSafeInstance() bool {
	return tc.safeInstance
}

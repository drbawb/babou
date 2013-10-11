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
	*App
	auth   *filters.AuthContext
	events *filters.EventContext

	actionMap map[string]web.Action
}

// Returns a routable instance of TorrentController
func NewTorrentController() *TorrentController {
	return &TorrentController{}
}

func (tc *TorrentController) Dispatch(action string) (web.Controller, web.Action) {
	newTc := &TorrentController{
		actionMap: make(map[string]web.Action),
		App:       &App{},
	}

	//add your actions here.
	newTc.actionMap["index"] = newTc.Index
	newTc.actionMap["latestEpisodes"] = newTc.LatestEpisodes

	newTc.actionMap["new"] = newTc.New
	newTc.actionMap["create"] = newTc.Create

	newTc.actionMap["download"] = newTc.Download

	newTc.actionMap["delete"] = newTc.Delete

	return newTc, newTc.actionMap[action]
}

func (tc *TorrentController) Index() *web.Result {
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

	for _, t := range torrentList {
		stats := tc.events.ReadStats(t.InfoHash)
		if stats == nil {
			continue
		}

		t.Seeding = stats.Seeding
		t.Leeching = stats.Leeching
	}

	outData.TorrentList = torrentList

	output.Body = []byte(web.RenderWith("bootstrap", "torrent", "index", outData, tc.Flash))

	return output
}

func (tc *TorrentController) LatestEpisodes() *web.Result {

	outData := &struct {
		SeriesList []*models.SeriesBundle
	}{
		SeriesList: models.LatestSeries(),
	}

	result := &web.Result{
		Status: 200,
		Body: []byte(web.RenderWith(
			"bootstrap",
			"torrent",
			"series",
			outData,
			tc.Flash)),
	}

	return result
}

// Displays a form where a user can upload a new torrent.
func (tc *TorrentController) New() *web.Result {
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
	output.Body = []byte(web.RenderWith("bootstrap", "torrent", "new", outData, tc.Flash))

	return output
}

func (tc *TorrentController) Create() *web.Result {
	//TODO: permission check; for now any authenticated user can add torrents.
	redirect, user := tc.RedirectOnAuthFail()
	if user == nil {
		return redirect
	}

	formFiles := tc.Dev.Params.Files
	if formFiles["metainfo"] == nil {
		tc.Flash.AddFlash("File upload appears to be missing.")
	} else if len(formFiles["metainfo"]) <= 0 || len(formFiles["metainfo"]) > 1 {
		tc.Flash.AddFlash("You are only allowed to upload one torrent at a time.")
	} else {
		file := formFiles["metainfo"][0]
		if !strings.HasSuffix(file.Filename, ".torrent") {
			tc.Flash.AddFlash("File needs to end w/ .torrent!")
			return tc.RedirectOnUploadFail()
		}

		fmt.Printf("reading torrent: \n")
		torrentFile, err := file.Open()
		if err != nil {
			tc.Flash.AddFlash("Error reading your torrent; please try your upload again")
			return tc.RedirectOnUploadFail()
		}

		torrent := libTorrent.ReadFile(torrentFile)
		torrentRecord := &models.Torrent{}

		if err := torrentRecord.Populate(torrent.Info); err != nil {
			tc.Flash.AddFlash("Error reading your torrent file.")
			return tc.RedirectOnUploadFail()
		}

		attributes := &models.Attribute{}
		attributes.AlbumName = tc.Dev.Params.All["albumName"]
		attributes.ReleaseYear = time.Now()
		attributes.ArtistName = strings.Split(tc.Dev.Params.All["artistName"], ",")
		fmt.Printf("num artists: %d \n", len(attributes.ArtistName))

		torrentRecord.SetAttributes(attributes)

		if err := torrentRecord.Write(); err != nil {
			tc.Flash.AddFlash(fmt.Sprintf("Error saving your torrent. Please contact a staff member: %s", err.Error()))
			return tc.RedirectOnUploadFail()
		}

		tc.Flash.AddFlash(fmt.Sprintf(`Your torrents URL is: http://tracker.fatalsyntax.com/torrents/download/%d -- 
			please save this because babou cannot find things right now.`, torrentRecord.ID))
	}

	// Issue redirect to index page.
	output := &web.Result{
		Redirect: &web.RedirectPath{
			NamedRoute: "torrentIndex",
		},
		Status: 302,
	}

	return output
}

func (tc *TorrentController) Download() *web.Result {
	redirect, user := tc.RedirectOnAuthFail()
	if user == nil {
		return redirect
	}

	output := &web.Result{
		Status: 200,
	}

	record := &models.Torrent{}

	var torrentId int64
	torrentId, err := strconv.ParseInt(tc.Dev.Params.All["torrentId"], 10, 32)

	if err != nil {
		output.Body = []byte("invalid torrent id.")
		return output
	}

	record.SelectId(int(torrentId))
	outFile, err := record.WriteFile(user.Secret, user.SecretHash)
	if err != nil {
		result := &web.Result{}
		tc.Flash.AddFlash("Could not find the torrent with the specified ID")
		result.Redirect = &web.RedirectPath{}
		result.Redirect.NamedRoute = "torrentIndex"

		result.Status = 302

		return result
	}

	output.IsFile = true
	output.Filename = record.Name + ".torrent"
	output.Body = outFile

	//TODO: Test attributes.
	attributes, err := record.Attributes()
	if attributes != nil {
		fmt.Printf("attributes: %v \n", attributes)
	}

	return output
}

func (tc *TorrentController) Delete() *web.Result {
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

// Setup contexts

func (tc *TorrentController) SetAuthContext(context *filters.AuthContext) error {
	tc.auth = context
	return nil
}

func (tc *TorrentController) SetEventContext(context *filters.EventContext) error {
	tc.events = context

	return nil
}

// Tests that the current chain is sufficient for this route.
func (tc *TorrentController) TestContext(chain []web.ChainableContext) error {
	err := tc.App.TestContext(chain)
	if err != nil {
		return err
	}

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

	if !outFlag {
		return errors.New("Auth chain missing from torrent route.")
	}

	if !eventFlag {
		return errors.New("Event chian missing from torrent route")
	}

	return nil
}

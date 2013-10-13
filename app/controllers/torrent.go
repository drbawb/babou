package controllers

import (
	filters "github.com/drbawb/babou/app/filters"
	models "github.com/drbawb/babou/app/models"

	libTorrent "github.com/drbawb/babou/lib/torrent"
	web "github.com/drbawb/babou/lib/web"

	"github.com/drbawb/babou/bridge"

	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type TorrentController struct {
	*App
	auth   *filters.AuthContext
	events *filters.EventContext

	actionMap    map[string]web.Action
	acceptHeader string
}

// Returns a routable instance of TorrentController
func NewTorrentController() *TorrentController {
	return &TorrentController{}
}

func (tc *TorrentController) Dispatch(action, accept string) (web.Controller, web.Action) {
	newTc := &TorrentController{
		actionMap:    make(map[string]web.Action),
		acceptHeader: accept,
		App:          &App{},
	}

	//add your actions here.
	newTc.actionMap["index"] = newTc.Index
	newTc.actionMap["latestEpisodes"] = newTc.Episodes
	newTc.actionMap["latestSeries"] = newTc.Series

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

func (tc *TorrentController) Episodes() *web.Result {

	outData := &struct {
		EpisodeList  []*models.EpisodeBundle
		ShowEpisodes bool
	}{
		EpisodeList:  models.LatestEpisodes(),
		ShowEpisodes: true,
	}

	// Respond with?
	result := &web.Result{Status: 200}

	if strings.Contains(tc.acceptHeader, "application/json") {
		// marshal
		jsonResponse, err := json.Marshal(outData.EpisodeList)
		if err != nil {
			result.Status = 500
			result.Body = []byte("error formatting json for resp.")
			return result
		}

		result.Body = jsonResponse
	} else {
		result.Body = []byte(web.RenderWith(
			"bootstrap",
			"torrent",
			"tv",
			outData,
			tc.Flash))
	}

	return result
}

func (tc *TorrentController) Series() *web.Result {
	outData := &struct {
		SeriesList []*models.SeriesBundle
		ShowSeries bool
	}{
		SeriesList: models.LatestSeries(),
		ShowSeries: true,
	}

	// Respond with?
	result := &web.Result{Status: 200}

	if strings.Contains(tc.acceptHeader, "application/json") {
		// marshal
		jsonResponse, err := json.Marshal(outData.SeriesList)
		if err != nil {
			result.Status = 500
			result.Body = []byte("error formatting json for resp.")
			return result
		}

		result.Body = jsonResponse
	} else {
		result.Body = []byte(web.RenderWith(
			"bootstrap",
			"torrent",
			"tv",
			outData,
			tc.Flash))
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

		if err := torrentRecord.Write(); err != nil {
			tc.Flash.AddFlash(fmt.Sprintf("Error saving your torrent. Please contact a staff member: %s", err.Error()))
			return tc.RedirectOnUploadFail()
		}

		// Write attributes bundle [TV]
		switch tc.Dev.Params.All["category"] {
		case "series":
			fmt.Printf("Series attributes bundle")
			sBundle := &models.SeriesBundle{}
			sBundle.Name = tc.Dev.Params.All["seriesName"] //TODO: frontend: autocomplete, backend: verify (try to avoid dups, basically.)

			if err := sBundle.Persist(); err != nil {
				fmt.Printf("Error saving bundle: %s", err.Error())
			}
		case "episode":
			fmt.Printf("Episode attributes bundle")
			// lookup sBundle by name
			sBundle := &models.SeriesBundle{}
			sBundle.SelectByName(tc.Dev.Params.All["seriesName"])

			eBundle := &models.EpisodeBundle{}
			eBundle.Name = tc.Dev.Params.All["episodeName"]
			eBundle.Format = tc.Dev.Params.All["format"]
			if eBundle.Number, err = strconv.Atoi(tc.Dev.Params.All["episodeNumber"]); err != nil {
				fmt.Printf("Error converting episode number to integer!")
			}

			if err := eBundle.PersistWithSeries(sBundle); err != nil {
				fmt.Printf("Error saving episode bundle: %s", err.Error())
			}

		default:
			fmt.Printf("No attributes bundle create")

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

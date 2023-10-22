package structs

import (
	Utils "github.com/COMTOP1/AFC-GO/myUtils"
	"github.com/COMTOP1/AFC-GO/objects"
)

type (
	PageParams struct {
		//Base     BaseParams
		MyUtils   Utils.MyUtils
		Document  objects.Documents
		Document1 objects.Documents
		Member    Member
		Team      []Member
	}
	Member struct {
		Name, Role string
	}

	// Config is a structure containing global website configuration.
	//
	// See the comments for Server and PageContext for more details.
	Config struct {
		Server      Server      `toml:"server"`
		PageContext PageContext `toml:"page_context"`
		Mail        Mail        `toml:"mail"`
	}

	// Server is a structure containing server configuration.
	Server struct {
		Debug            bool   `toml:"debug"`
		Port             string `toml:"port"`
		DomainName       string `toml:"domain_name"`
		Version          string `toml:"version"`
		Commit           string `toml:"commit"`
		BSWDI_API        string `toml:"bswdi_api"`
		AccessCookieName string `toml:"access_cookie_name"`
	}

	// PageContext is a structure containing static information to provide
	// to all page templates.
	//
	// This contains the website's long and short names, as well as a directory
	// of pages for navigation.
	PageContext struct {
		//LongName         string                `toml:"longName"`
		//ShortName        string                `toml:"shortName"`
		//SiteDescription  string                `toml:"siteDescription"`
		URLPrefix string `toml:"url_prefix"`
		FullURL   string `toml:"full_url"`
		//MainTwitter      string                `toml:"mainTwitter"`
		//MainFacebook     string                `toml:"mainFacebook"`
		//MainInstagram    string                `toml:"mainInstagram"`
		//NewsTwitter      string                `toml:"newsTwitter"`
		MyRadioAPIKey string `toml:"public_my_radio_api_key"`
		//ODName           string                `toml:"odName"`
		//Christmas        bool                  `toml:"christmas"`
		//AprilFools       bool                  `toml:"aprilFools"`
		//CIN              bool                  `toml:"cin"`
		//CINLivestreaming bool                  `toml:"cinLivestreaming"`
		//CINAPI           string                `toml:"cinAPI"`
		//CINHashtag       string                `toml:"cinHashtag"`
		//CINLive          string                `toml:"cinLive"`
		//IndexCountdown *IndexCountdownConfig `toml:"indexCountdown"`
		//CacheBuster    string                `toml:"cacheBuster"`
		//Pages          []Page
		//Youtube        youtube
		//Gmaps          gmaps
		GetTime     func() int64
		GetTime1Day func() int64
		GetDay      func(Time int64) string
		GetYear     func() int
	}

	Mail struct {
		Enabled  bool   `toml:"mail_enabled"`
		Host     string `toml:"mail_host"`
		User     string `toml:"mail_user"`
		Password string `toml:"mail_pass"`
		Port     int    `toml:"mail_port"`
	}
)

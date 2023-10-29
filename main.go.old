package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/COMTOP1/AFC-GO/controllers"
	"github.com/COMTOP1/AFC-GO/routes"
	"github.com/COMTOP1/AFC-GO/structs"
	"github.com/COMTOP1/AFC-GO/utils"
	"github.com/COMTOP1/api/handler"
	"html/template"
	"log"
	"os"
	"os/signal"
)

func main() {
	//router := echo.New()
	//_ = router
	/*year, month, day := time.Now().Date()
	fmt.Println(myUtils.GetTime())
	fmt.Printf("%s %d %s %d\n", time.Now().Weekday().String()[:3], day, month.String()[:3], year)
	fmt.Println(Role.FindRole("LeagueSecretary"))

	role := objects.Role{Role: Role.ProgrammeEditor}

	fmt.Println(role.ToString())

	role.SetRole(Role.FindRole("ClubSecretary"))

	fmt.Println(role)
	fmt.Println(role.ToString())

	roleCopy := role

	roleCopy.SetRole(Role.Webmaster)

	fmt.Println(role)
	fmt.Println(roleCopy)

	//return

	aff := objects.Affiliations{
		Id:      1,
		Name:    "Test01",
		Website: "http://test.01.com",
		Image:   "null",
	}

	fmt.Println(aff)*/

	var err error

	config := &structs.Config{}
	_, err = toml.DecodeFile("config.toml", config)
	if err != nil {
		log.Fatal(err)
	}

	if config.Server.Debug {
		log.SetFlags(log.Llongfile)
	}

	access := utils.NewAccesser(utils.Config{
		AccessCookieName: config.Server.AccessCookieName, // jwt_token --> base64
		DomainName:       config.Server.DomainName,
	})

	var mailer *utils.Mailer
	if config.Mail.Host != "" {
		if config.Mail.Enabled {
			mailConfig := utils.MailConfig{
				Host:     config.Mail.Host,
				Port:     config.Mail.Port,
				Username: config.Mail.User,
				Password: config.Mail.Password,
			}

			mailer, err = utils.NewMailer(mailConfig)
			if err != nil {
				log.Printf("failed to connect to mail server: %+v", err)
				config.Mail.Enabled = false
			} else {
				log.Printf("Connected to mail server: %s\n", config.Mail.Host)

				mailer.Defaults = utils.Defaults{
					DefaultTo:   "root@bswdi.co.uk",
					DefaultFrom: "BSWDI AFC <afc@bswdi.co.uk>",
				}
			}
		}
	} else {
		config.Mail.Enabled = false
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if config.Mail.Enabled {
				exitingTemplate := template.New("Exiting Template")
				exitingTemplate = template.Must(exitingTemplate.Parse("<body>BSWDI AFC has been stopped!<br><br>{{if .Debug}}Exit signal: {{.Sig}}<br><br>{{end}}Version: {{.Version}}<br>Commit: {{.Commit}}</body>"))

				starting := utils.Mail{
					Subject:     "BSWDI AFC has been stopped!",
					UseDefaults: true,
					Tpl:         exitingTemplate,
					TplData: struct {
						Debug           bool
						Sig             os.Signal
						Version, Commit string
					}{
						Debug:   config.Server.Debug,
						Sig:     sig,
						Version: config.Server.Version,
						Commit:  config.Server.Commit,
					},
				}

				err = mailer.SendMail(starting)
				if err != nil {
					fmt.Println(err)
				}
				err = mailer.Close()
				if err != nil {
					fmt.Println(err)
				}
			}
			os.Exit(0)
		}
	}()

	//err = routes.New(config, access, mailer).Start()
	if err != nil {
		if mailer != nil {
			err1 := mailer.SendErrorFatalMail(utils.Mail{
				Error:       fmt.Errorf("the web server couldn't be started: %s... exiting", err),
				UseDefaults: true,
			})
			if err1 != nil {
				fmt.Println(err1)
			}
		}
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}
	if err != nil {
		if mailer != nil {
			err1 := mailer.SendErrorFatalMail(utils.Mail{
				Error:       fmt.Errorf("the web server couldn't be started: %s... exiting", err),
				UseDefaults: true,
			})
			if err1 != nil {
				fmt.Println(err1)
			}
		}
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}

	session, err := handler.NewSession(config.Server.BSWDI_API)
	if err != nil {
		if config.Mail.Enabled {
			err1 := mailer.SendErrorFatalMail(utils.Mail{
				Error:       fmt.Errorf("the session couldn't be initialised: %s... exiting", err),
				UseDefaults: true,
			})
			if err1 != nil {
				fmt.Println(err1)
			}
		}
		log.Fatalf("The session couldn't be initialised!\n\n%s\n\nExiting!", err)
	}

	log.Printf("AFC Version %s", config.Server.Version)

	if config.Mail.Enabled {

		startingTemplate := template.New("Starting Template")
		startingTemplate = template.Must(startingTemplate.Parse("<body>BSWDI AFC starting{{if .Debug}} in debug mode!<br><b>Do not run in production! Authentication is disabled!</b>{{else}}!{{end}}<br><br>Version: {{.Version}}<br>Commit: {{.Commit}}<br><br>If you don't get another email then this has started correctly.</body>"))

		subject := "BSWDI AFC is starting"

		if config.Server.Debug {
			subject += " in debug mode"
			log.Println("Debug Mode - Disabled auth - do not run in production!")
		}

		subject += "!"

		starting := utils.Mail{
			Subject:     subject,
			UseDefaults: true,
			Tpl:         startingTemplate,
			TplData: struct {
				Debug           bool
				Version, Commit string
			}{
				Debug:   config.Server.Debug,
				Version: config.Server.Version,
				Commit:  config.Server.Commit,
			},
		}

		err = mailer.SendMail(starting)
		if err != nil {
			fmt.Println(err)
		}
	}

	controller := controllers.GetController(access, session)

	router1 := routes.New(&routes.NewRouter{
		Config:   config,
		Port:     config.Server.Port,
		Accesser: access,
		Repos:    controllers.NewRepos(controller),
		//DomainName:
		Debug: config.Server.Debug,
		//Access: accesser,
		Mailer: mailer,
	})
	err = router1.Start()
	if err != nil {
		if config.Mail.Enabled {
			err1 := mailer.SendErrorFatalMail(utils.Mail{
				Error:       fmt.Errorf("the web server couldn't be started: %s... exiting", err),
				UseDefaults: true,
			})
			if err1 != nil {
				fmt.Println(err1)
			}
		}
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}

	//web := Web{mux: mux.NewRouter()}
	//log.Println("Web loaded")
	////web.team, err = team.New()
	//if err != nil {
	//	log.Printf("failed to get team: %+v\n", err)
	//}
	//web.mux.HandleFunc("/set", setCookieHandler)
	//web.mux.HandleFunc("/get", getCookieHandler)
	//web.mux.HandleFunc("/", web.home)
	//web.mux.HandleFunc("/home", web.home)
	//web.mux.HandleFunc("/public/webfonts/Arial/{id:[a-zA-Z0-9_.-]+}", web.publicFontArial)
	//web.mux.HandleFunc("/public/webfonts/Allerta/{id:[a-zA-Z0-9_.-]+}", web.publicFontAllerta)
	//web.mux.HandleFunc("/public/{id:[a-zA-Z0-9_.-]+}", web.public)
	//log.Println("AFC site: 0.0.0.0:7075")
	//log.Fatal(http.ListenAndServe("0.0.0.0:7075", web.mux))
}

//func (web *Web) home(w http.ResponseWriter, _ *http.Request) {
//	if verbose {
//		log.Println("Home called")
//	}
//
//	//config := &structs.Config{}
//
//	//config.PageContext.CurrentYear = time.Now().Year()
//
//	//web.t = templates.NewMain()
//
//	a := make([]structs.Member, 4)
//	a[0] = structs.Member{
//		Name: "Liam Burnand",
//		Role: "Computing Director",
//	}
//	a[1] = structs.Member{
//		Name: "Dan Wade",
//		Role: "Deputy Computing Director",
//	}
//	a[2] = structs.Member{
//		Name: "Rhys Milling",
//	}
//	a[3] = structs.Member{
//		Name: "Marks Polakovs",
//	}
//
//	session, err := handler.NewSession("http://localhost:8081/bb2f24a2dd9832c60c3e5d5a3cc161c081dad378/v1/")
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	token := "1eyJhbGciOiJIUzUxMiIsImtpZCI6ImZmNzA0NzczLTczODQtNDc0My1iODY1LTYxOWI3MTIxOTFhZCIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiV2VibWFzdGVyIiwiaWQiOjEsImF1ZCI6Imh0dHBzOi8vYWZjYWxkZXJtYXN0b24uY28udWsiLCJleHAiOjE2NzYwMzE3MDgsImp0aSI6IjE1ZDg0NGI1LWM2NTMtNDM1Ni1hYzE5LTU3YzBlODhiZmUxMCIsImlhdCI6MTY3NjAzMDUwOCwiaXNzIjoiaHR0cHM6Ly9zc28uYnN3ZGkuY28udWsiLCJuYmYiOjE2NzYwMzA1MDh9.7buFzCfPvYGmY5tFsYZpYiKsPYKJXxNafmcvm8huJxAD4FtS_wOI_8Yd1o3OrAJwYLEPDznETujr8C5my6tXqg"
//
//	user, err := session.GetUserByToken(token)
//	if err != nil {
//		if !strings.Contains(err.Error(), "invalid token") {
//			fmt.Println(err)
//		}
//	}
//
//	page := structs.PageParams{
//		//Base: structs.BaseParams{
//		//	SystemTime: time.Now(),
//		//},
//		MyUtils: myUtils.MyUtils{},
//		Document: objects.Documents{
//			Id:       6547,
//			Name:     "Hello",
//			FileName: "This",
//		},
//		Team: a,
//	}
//
//	affiliations, err := session.ListAllAffiliations()
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	sponsors, err := session.ListAllSponsorsMinimal()
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	newsLatest, err := session.GetNewsLatest()
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	whatsOnLatest, err := session.GetWhatsOnLatest()
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	data := struct {
//		GetTest3 func() string
//		MyUtils  struct {
//			GetTest  func() string
//			GetTest2 func() string
//			GetYear  func() int
//		}
//		GetTime       func() int64
//		GetDate       func(Time int64) string
//		Affiliations  []handler.Affiliation
//		Sponsors      []handler.Sponsor
//		NewsLatest    handler.News
//		WhatsOnLatest handler.WhatsOn
//		User          handler.User
//		GetTeamName   func(id uint64) string
//	}{
//		GetTest3: func() string {
//			return "test"
//		},
//		MyUtils: struct {
//			GetTest  func() string
//			GetTest2 func() string
//			GetYear  func() int
//		}{
//			GetTest:  page.MyUtils.GetTest,
//			GetTest2: page.MyUtils.GetTest2,
//			GetYear:  page.MyUtils.GetYear,
//		},
//		GetTime:       page.MyUtils.GetTime,
//		GetDate:       page.MyUtils.GetDay,
//		Affiliations:  affiliations,
//		Sponsors:      sponsors,
//		NewsLatest:    newsLatest,
//		WhatsOnLatest: whatsOnLatest,
//		User:          user,
//		GetTeamName: func(id uint64) string {
//			team, err := session.GetTeamById(id)
//			if err != nil {
//				fmt.Println(err)
//				return "TEAM NOT FOUND!"
//			}
//			return team.Name
//		},
//	}
//
//	err = web.t.RenderTemplate(w, page, data, "home.tmpl")
//	//err = web.t.Page(w, page)
//	if err != nil {
//		err = fmt.Errorf("failed to render dashboard: %w", err)
//		fmt.Println(err)
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//	}
//}
//
//// public is the handler for any public documents, for example, the style sheet or images
//func (web *Web) public(w http.ResponseWriter, r *http.Request) {
//	if verbose {
//		log.Println("Public called")
//	}
//	vars := mux.Vars(r)
//	http.ServeFile(w, r, "public/"+vars["id"])
//}
//
//func (web *Web) publicFontArial(w http.ResponseWriter, r *http.Request) {
//	if verbose {
//		log.Println("Public Font Arial called")
//	}
//	vars := mux.Vars(r)
//	http.ServeFile(w, r, "public/webfonts/Arial/"+vars["id"])
//}
//
//func (web *Web) publicFontAllerta(w http.ResponseWriter, r *http.Request) {
//	if verbose {
//		log.Println("Public Font Allerta called")
//	}
//	vars := mux.Vars(r)
//	http.ServeFile(w, r, "public/webfonts/Allerta/"+vars["id"])
//}

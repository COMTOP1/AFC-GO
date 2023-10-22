package controllers

type Repos struct {
	Account *AccountRepo
	//Affiliations     *affiliations.Repo
	//Documents        *documents.Repo
	//Images           *images.Repo
	Download *DownloadRepo
	Home     *HomeRepo
	Public   *PublicRepo
	//News             *news.Repo
	//Players          *players.Repo
	//Programmes       *programmes.Repo
	//ProgrammeSeasons *programmeSeasons.Repo
	//Sponsors         *sponsors.Repo
	Teams *TeamsRepo
	//Users            *users.Repo
	//WhatsOn          *whatsOn.Repo
}

func NewRepos(controller Controller) *Repos {
	return &Repos{
		Account: NewAccountRepo(controller),
		//Affiliations:     affiliations.NewRepo(controller),
		//Documents:        documents.NewRepo(controller),
		//Images:           images.NewRepo(controller),
		Download: NewDownloadRepo(controller),
		Home:     NewHomeRepo(controller),
		Public:   NewPublicRepo(controller),
		//News:             news.NewRepo(controller),
		//Players:          players.NewRepo(controller),
		//Programmes:       programmes.NewRepo(controller),
		//ProgrammeSeasons: programmeSeasons.NewRepo(controller),
		//Sponsors:         sponsors.NewRepo(controller),
		Teams: NewTeamsRepo(controller),
		//Users:            users.NewRepo(controller),
		//WhatsOn:          whatsOn.NewRepo(controller),
	}
}

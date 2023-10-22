package objects

type (
	Teams struct {
		Id, Ages                                                                uint64
		Name, League, Division, LeagueTable, Fixtures, Coach, Physio, TeamPhoto string
		Active, Youth                                                           bool
	}
)

func (t *Teams) SetFields(Id, Ages uint64, Name, League, Division, LeagueTable, Fixtures, Coach, Physio, TeamPhoto string, Active, Youth bool) {
	t.Id = Id
	t.Ages = Ages
	t.Name = Name
	t.League = League
	t.Division = Division
	t.LeagueTable = LeagueTable
	t.Fixtures = Fixtures
	t.Coach = Coach
	t.Physio = Physio
	t.TeamPhoto = TeamPhoto
	t.Active = Active
	t.Youth = Youth
}

func (t *Teams) SetFieldsNotImage(Id, Ages uint64, Name, League, Division, LeagueTable, Fixtures, Coach, Physio string, Active, Youth bool) {
	t.Id = Id
	t.Ages = Ages
	t.Name = Name
	t.League = League
	t.Division = Division
	t.LeagueTable = LeagueTable
	t.Fixtures = Fixtures
	t.Coach = Coach
	t.Physio = Physio
	t.Active = Active
	t.Youth = Youth
}

func (t *Teams) GetId() uint64 {
	return t.Id
}

func (t *Teams) SetId(Id uint64) {
	t.Id = Id
}

func (t *Teams) GetAges() uint64 {
	return t.Ages
}

func (t *Teams) SetAges(Ages uint64) {
	t.Ages = Ages
}

func (t *Teams) GetName() string {
	return t.Name
}

func (t *Teams) SetName(Name string) {
	t.Name = Name
}

func (t *Teams) GetLeague() string {
	return t.League
}

func (t *Teams) SetLeague(League string) {
	t.League = League
}

func (t *Teams) GetDivision() string {
	return t.Division
}

func (t *Teams) SetDivision(Division string) {
	t.Division = Division
}

func (t *Teams) GetLeagueTable() string {
	return t.LeagueTable
}

func (t *Teams) SetLeagueTable(LeagueTable string) {
	t.LeagueTable = LeagueTable
}

func (t *Teams) GetFixtures() string {
	return t.Fixtures
}

func (t *Teams) SetFixtures(Fixtures string) {
	t.Fixtures = Fixtures
}

func (t *Teams) GetCoach() string {
	return t.Coach
}

func (t *Teams) SetCoach(Coach string) {
	t.Coach = Coach
}

func (t *Teams) GetPhysio() string {
	return t.Physio
}

func (t *Teams) SetPhysio(Physio string) {
	t.Physio = Physio
}

func (t *Teams) GetTeamPhoto() string {
	return t.TeamPhoto
}

func (t *Teams) SetTeamPhoto(TeamPhoto string) {
	t.TeamPhoto = TeamPhoto
}

func (t *Teams) GetActive() bool {
	return t.Active
}

func (t *Teams) SetActive(Active bool) {
	t.Active = Active
}

func (t *Teams) GetYouth() bool {
	return t.Youth
}

func (t *Teams) SetYouth(Youth bool) {
	t.Youth = Youth
}

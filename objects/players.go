package objects

type (
	Players struct {
		Id                    uint64
		TeamId, DateOfBirth   int64
		Name, Image, Position string
		Captain               bool
	}
)

func (p *Players) SetFields(Id uint64, TeamId, DateOfBirth int64, Name, Image, Position string, Captain bool) {
	p.Id = Id
	p.TeamId = TeamId
	p.DateOfBirth = DateOfBirth
	p.Name = Name
	p.Image = Image
	p.Position = Position
	p.Captain = Captain
}

func (p *Players) SetFieldsNotImage(Id uint64, TeamId, DateOfBirth int64, Name, Position string, Captain bool) {
	p.Id = Id
	p.TeamId = TeamId
	p.DateOfBirth = DateOfBirth
	p.Name = Name
	p.Position = Position
	p.Captain = Captain
}

func (p *Players) GetId() uint64 {
	return p.Id
}

func (p *Players) SetId(Id uint64) {
	p.Id = Id
}

func (p *Players) GetTeamId() int64 {
	return p.TeamId
}

func (p *Players) SetTeamId(TeamId int64) {
	p.TeamId = TeamId
}

func (p *Players) GetDateOfBirth() int64 {
	return p.DateOfBirth
}

func (p *Players) SetDateOfBirth(DateOfBirth int64) {
	p.DateOfBirth = DateOfBirth
}

func (p *Players) GetName() string {
	return p.Name
}

func (p *Players) SetName(Name string) {
	p.Name = Name
}

func (p *Players) GetImage() string {
	return p.Image
}

func (p *Players) SetImage(Image string) {
	p.Image = Image
}

func (p *Players) GetPosition() string {
	return p.Position
}

func (p *Players) SetPosition(Position string) {
	p.Position = Position
}

func (p *Players) GetCaptain() bool {
	return p.Captain
}

func (p *Players) SetCaptain(Captain bool) {
	p.Captain = Captain
}

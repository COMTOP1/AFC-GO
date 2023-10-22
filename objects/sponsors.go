package objects

type (
	Sponsors struct {
		Id                                    uint64
		Name, Website, Image, Purpose, TeamId string
	}
)

func (s *Sponsors) SetFields(Id uint64, Name, Website, Image, Purpose, TeamId string) {
	s.Id = Id
	s.Name = Name
	s.Website = Website
	s.Image = Image
	s.Purpose = Purpose
	s.TeamId = TeamId
}

func (s *Sponsors) SetFieldsNotImage(Id uint64, Name, Website, Purpose, TeamId string) {
	s.Id = Id
	s.Name = Name
	s.Website = Website
	s.Purpose = Purpose
	s.TeamId = TeamId
}

func (s *Sponsors) GetId() uint64 {
	return s.Id
}

func (s *Sponsors) SetId(Id uint64) {
	s.Id = Id
}

func (s *Sponsors) GetName() string {
	return s.Name
}

func (s *Sponsors) SetName(Name string) {
	s.Name = Name
}

func (s *Sponsors) GetWebsite() string {
	return s.Website
}

func (s *Sponsors) SetWebsite(Website string) {
	s.Website = Website
}

func (s *Sponsors) GetImage() string {
	return s.Image
}

func (s *Sponsors) SetImage(Image string) {
	s.Image = Image
}

func (s *Sponsors) GetPurpose() string {
	return s.Purpose
}

func (s *Sponsors) SetPurpose(Purpose string) {
	s.Purpose = Purpose
}

func (s *Sponsors) GetTeamId() string {
	return s.TeamId
}

func (s *Sponsors) SetTeamId(TeamId string) {
	s.TeamId = TeamId
}

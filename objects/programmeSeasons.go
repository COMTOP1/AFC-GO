package objects

type (
	ProgrammeSeasons struct {
		Id     uint64
		Season string
	}
)

func (p *ProgrammeSeasons) SetFields(Id uint64, Season string) {
	p.Id = Id
	p.Season = Season
}

func (p *ProgrammeSeasons) GetId() uint64 {
	return p.Id
}

func (p *ProgrammeSeasons) SetId(Id uint64) {
	p.Id = Id
}

func (p *ProgrammeSeasons) GetSeason() string {
	return p.Season
}

func (p *ProgrammeSeasons) SetSeason(Season string) {
	p.Season = Season
}

package objects

type (
	Programmes struct {
		Id, ProgrammeSeasonId uint64
		Name, FileName        string
		DateOfProgramme       int64
	}
)

func (p *Programmes) SetFields(Id, ProgrammeSeasonId uint64, Name, FileName string, DateOfProgramme int64) {
	p.Id = Id
	p.ProgrammeSeasonId = ProgrammeSeasonId
	p.Name = Name
	p.FileName = FileName
	p.DateOfProgramme = DateOfProgramme
}

func (p *Programmes) GetId() uint64 {
	return p.Id
}

func (p *Programmes) SetId(Id uint64) {
	p.Id = Id
}

func (p *Programmes) GetProgrammeSeasonId() uint64 {
	return p.ProgrammeSeasonId
}

func (p *Programmes) SetProgrammeSeasonId(ProgrammeSeasonId uint64) {
	p.ProgrammeSeasonId = ProgrammeSeasonId
}

func (p *Programmes) GetName() string {
	return p.Name
}

func (p *Programmes) SetName(Name string) {
	p.Name = Name
}

func (p *Programmes) GetFileName() string {
	return p.FileName
}

func (p *Programmes) SetFileName(FileName string) {
	p.FileName = FileName
}

func (p *Programmes) GetDateOfProgramme() int64 {
	return p.DateOfProgramme
}

func (p *Programmes) SetDateOfProgramme(DateOfProgramme int64) {
	p.DateOfProgramme = DateOfProgramme
}

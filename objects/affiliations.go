package objects

type (
	Affiliations struct {
		Id                   uint64
		Name, Website, Image string
	}
)

func (a *Affiliations) SetFields(Id uint64, Name, Website, Image string) {
	a.Id = Id
	a.Name = Name
	a.Website = Website
	a.Image = Image
}

func (a *Affiliations) SetFieldsNotImage(Id uint64, Name, Website string) {
	a.Id = Id
	a.Name = Name
	a.Website = Website
}

func (a *Affiliations) GetId() uint64 {
	return a.Id
}

func (a *Affiliations) SetId(Id uint64) {
	a.Id = Id
}

func (a *Affiliations) GetName() string {
	return a.Name
}

func (a *Affiliations) SetName(Name string) {
	a.Name = Name
}

func (a *Affiliations) GetWebsite() string {
	return a.Website
}

func (a *Affiliations) SetWebsite(Website string) {
	a.Website = Website
}

func (a *Affiliations) GetImage() string {
	return a.Image
}

func (a *Affiliations) SetImage(Image string) {
	a.Image = Image
}

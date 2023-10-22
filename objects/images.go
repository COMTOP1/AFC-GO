package objects

type (
	Images struct {
		Id             uint64
		Image, Caption string
	}
)

func (i *Images) SetFields(Id uint64, Image, Caption string) {
	i.Id = Id
	i.Image = Image
	i.Caption = Caption
}

func (i *Images) SetFieldsNotImage(Id uint64, Caption string) {
	i.Id = Id
	i.Caption = Caption
}

func (i *Images) GetId() uint64 {
	return i.Id
}

func (i *Images) SetId(Id uint64) {
	i.Id = Id
}

func (i *Images) GetImage() string {
	return i.Image
}

func (i *Images) SetImage(Image string) {
	i.Image = Image
}

func (i *Images) GetCaption() string {
	return i.Caption
}

func (i *Images) SetCaption(Caption string) {
	i.Caption = Caption
}

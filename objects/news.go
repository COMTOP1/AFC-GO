package objects

type (
	News struct {
		Id                    uint64
		Title, Image, Content string
		Date                  int64
	}
)

func (n *News) SetFields(Id uint64, Title, Image, Content string, Date int64) {
	n.Id = Id
	n.Title = Title
	n.Image = Image
	n.Content = Content
	n.Date = Date
}

func (n *News) SetFieldsNotImage(Id uint64, Title, Content string, Date int64) {
	n.Id = Id
	n.Title = Title
	n.Content = Content
	n.Date = Date
}

func (n *News) GetId() uint64 {
	return n.Id
}

func (n *News) SetId(Id uint64) {
	n.Id = Id
}

func (n *News) GetTitle() string {
	return n.Title
}

func (n *News) SetTitle(Title string) {
	n.Title = Title
}

func (n *News) GetImage() string {
	return n.Image
}

func (n *News) SetImage(Image string) {
	n.Image = Image
}

func (n *News) GetContent() string {
	return n.Content
}

func (n *News) SetContent(Content string) {
	n.Content = Content
}

func (n *News) GetDate() int64 {
	return n.Date
}

func (n *News) SetDate(Date int64) {
	n.Date = Date
}

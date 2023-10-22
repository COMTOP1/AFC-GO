package objects

type (
	WhatsOn struct {
		Id                    uint64
		Title, Image, Content string
		Date, DateOfEvent     int64
	}
)

func (w *WhatsOn) SetFields(Id uint64, Title, Image, Content string, Date, DateOfEvent int64) {
	w.Id = Id
	w.Title = Title
	w.Image = Image
	w.Content = Content
	w.Date = Date
	w.DateOfEvent = DateOfEvent
}

func (w *WhatsOn) SetFieldsNotImage(Id uint64, Title, Content string, Date, DateOfEvent int64) {
	w.Id = Id
	w.Title = Title
	w.Content = Content
	w.Date = Date
	w.DateOfEvent = DateOfEvent
}

func (w *WhatsOn) GetId() uint64 {
	return w.Id
}

func (w *WhatsOn) SetId(Id uint64) {
	w.Id = Id
}

func (w *WhatsOn) GetTitle() string {
	return w.Title
}

func (w *WhatsOn) SetTitle(Title string) {
	w.Title = Title
}

func (w *WhatsOn) GetImage() string {
	return w.Image
}

func (w *WhatsOn) SetImage(Image string) {
	w.Image = Image
}

func (w *WhatsOn) GetContent() string {
	return w.Content
}

func (w *WhatsOn) SetContent(Content string) {
	w.Content = Content
}

func (w *WhatsOn) GetDate() int64 {
	return w.Date
}

func (w *WhatsOn) SetDate(Date int64) {
	w.Date = Date
}

func (w *WhatsOn) GetDateOfEvent() int64 {
	return w.DateOfEvent
}

func (w *WhatsOn) SetDateOfEvent(DateOfEvent int64) {
	w.DateOfEvent = DateOfEvent
}

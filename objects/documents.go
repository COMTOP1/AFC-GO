package objects

type (
	Documents struct {
		Id       uint64 `json:"id"`
		Name     string `json:"name"`
		FileName string `json:"file_name"`
	}
)

func (d *Documents) SetFields(Id uint64, Name, FileName string) {
	d.Id = Id
	d.Name = Name
	d.FileName = FileName
}

func (d *Documents) GetId() uint64 {
	return d.Id
}

func (d *Documents) SetId(Id uint64) {
	d.Id = Id
}

func (d *Documents) GetName() string {
	return d.Name
}

func (d *Documents) SetName(Name string) {
	d.Name = Name
}

func (d *Documents) GetFileName() string {
	return d.FileName
}

func (d *Documents) SetFileName(FileName string) {
	d.FileName = FileName
}

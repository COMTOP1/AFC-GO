package objects

type (
	Users struct {
		Name, Email, Image, Phone string
		TeamId                    uint64
		Role                      Role
	}
)

func (u *Users) SetFields(Name, Email, Image, Phone string, TeamId uint64, Role Role) {
	u.Name = Name
	u.Email = Email
	u.Image = Image
	u.Phone = Phone
	u.TeamId = TeamId
	u.Role = Role
}

func (u *Users) SetFieldsContacts(Name, Email, Phone string) {
	u.Name = Name
	u.Email = Email
	u.Phone = Phone
}

func (u *Users) GetName() string {
	return u.Name
}

func (u *Users) SetName(Name string) {
	u.Name = Name
}

func (u *Users) GetEmail() string {
	return u.Email
}

func (u *Users) SetEmail(Email string) {
	u.Email = Email
}

func (u *Users) GetImage() string {
	return u.Image
}

func (u *Users) SetImage(Image string) {
	u.Image = Image
}

func (u *Users) GetPhone() string {
	return u.Phone
}

func (u *Users) SetPhone(Phone string) {
	u.Phone = Phone
}

func (u *Users) GetTeamId() uint64 {
	return u.TeamId
}

func (u *Users) SetTeamId(TeamId uint64) {
	u.TeamId = TeamId
}

func (u *Users) GetRole() Role {
	return u.Role
}

func (u *Users) SetRole(Role Role) {
	u.Role = Role
}

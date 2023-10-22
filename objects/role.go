package objects

import "github.com/COMTOP1/AFC-GO/objects/role"

type (
	Role struct {
		role.Role
	}
)

func (r *Role) GetRole() string {
	switch r.Role {
	case role.Manager:
		return "Manager"
	case role.ProgrammeEditor:
		return "ProgrammeEditor"
	case role.LeagueSecretary:
		return "LeagueSecretary"
	case role.Treasurer:
		return "Treasurer"
	case role.SafeguardingOfficer:
		return "SafeguardingOfficer"
	case role.ClubSecretary:
		return "ClubSecretary"
	case role.Chairperson:
		return "Chairperson"
	case role.Webmaster:
		return "Webmaster"
	default:
		return ""
	}
}

func (r *Role) SetRole(Role role.Role) {
	r.Role = Role
}

func (r *Role) ToString() string {
	switch r.Role {
	case role.Manager:
		return "Manager"
	case role.ProgrammeEditor:
		return "Programme Editor"
	case role.LeagueSecretary:
		return "League Secretary"
	case role.Treasurer:
		return "Treasurer"
	case role.SafeguardingOfficer:
		return "Safeguarding Officer"
	case role.ClubSecretary:
		return "Club Secretary"
	case role.Chairperson:
		return "Chairperson"
	case role.Webmaster:
		return "Webmaster"
	default:
		return ""
	}
}

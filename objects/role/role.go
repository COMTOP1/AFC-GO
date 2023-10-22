package role

type (
	Role string
)

const (
	Manager             Role = "Manager"
	ProgrammeEditor     Role = "ProgrammeEditor"
	LeagueSecretary     Role = "LeagueSecretary"
	Treasurer           Role = "Treasurer"
	SafeguardingOfficer Role = "SafeguardingOfficer"
	ClubSecretary       Role = "ClubSecretary"
	Chairperson         Role = "Chairperson"
	Webmaster           Role = "Webmaster"
)

func FindRole(Role string) Role {
	switch Role {
	case "Manager":
		return Manager
	case "ProgrammeEditor":
		return ProgrammeEditor
	case "LeagueSecretary":
		return LeagueSecretary
	case "Treasurer":
		return Treasurer
	case "SafeguardingOfficer":
		return SafeguardingOfficer
	case "ClubSecretary":
		return ClubSecretary
	case "Chairperson":
		return Chairperson
	case "Webmaster":
		return Webmaster
	default:
		return ""
	}
}

package role

import (
	"fmt"
	"strings"
)

type Role string

const (
	Photographer        Role = "Photographer"
	Manager             Role = "Manager"
	ProgrammeEditor     Role = "Programme Editor"
	LeagueSecretary     Role = "League Secretary"
	Treasurer           Role = "Treasurer"
	SafeguardingOfficer Role = "Safeguarding Officer"
	ClubSecretary       Role = "Club Secretary"
	Chairperson         Role = "Chairperson"
	Webmaster           Role = "Webmaster"
)

func (r Role) DBString() string {
	switch r {
	case Photographer:
		return "PHOTOGRAPHER"
	case Manager:
		return "MANAGER"
	case ProgrammeEditor:
		return "PROGRAMME_EDITOR"
	case LeagueSecretary:
		return "LEAGUE_SECRETARY"
	case Treasurer:
		return "TREASURER"
	case SafeguardingOfficer:
		return "SAFEGUARDING_OFFICER"
	case ClubSecretary:
		return "CLUB_SECRETARY"
	case Chairperson:
		return "CHAIRPERSON"
	case Webmaster:
		return "WEBMASTER"
	default:
		return ""
	}
}

func GetRole(role string) (Role, error) {
	switch strings.ToLower(role) {
	case "photographer":
		return Photographer, nil
	case "manager":
		return Manager, nil
	case "programme_editor":
		return ProgrammeEditor, nil
	case "league_secretary":
		return LeagueSecretary, nil
	case "treasurer":
		return Treasurer, nil
	case "safeguarding_officer":
		return SafeguardingOfficer, nil
	case "club_secretary":
		return ClubSecretary, nil
	case "chairperson":
		return Chairperson, nil
	case "webmaster":
		return Webmaster, nil
	default:
		return "", fmt.Errorf("invalid role input: %s", role)
	}
}

func (r Role) String() string {
	return string(r)
}

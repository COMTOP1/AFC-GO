package main

import (
	"encoding/json"
	"fmt"
	"github.com/COMTOP1/api/handler"
	"io/ioutil"
	"log"
)

func main() {
	f, err := ioutil.ReadFile("json/teams.json")
	if err != nil {
		log.Fatal(err)
	}

	var teams []handler.Team

	err = json.Unmarshal(f, &teams)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(teams)

	session, err := handler.NewSession("http://localhost:8081/bb2f24a2dd9832c60c3e5d5a3cc161c081dad378/v1/")
	if err != nil {
		fmt.Println(err)
	}

	token := "eyJhbGciOiJIUzUxMiIsImtpZCI6IjhjY2VkNmQyLTcyY2YtNGFkNy05NDQ1LWFiMGRjOTZiMjgyNiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiV2VibWFzdGVyIiwiaWQiOjEsImF1ZCI6Imh0dHBzOi8vYWZjYWxkZXJtYXN0b24uY28udWsiLCJleHAiOjE2NzYyNTUzMTYsImp0aSI6IjJhN2JhOTcwLWU5ZTktNDFjYS1hODdkLTVmYWE4NDZiNTU3ZCIsImlhdCI6MTY3NjI1NDExNiwiaXNzIjoiaHR0cHM6Ly9zc28uYnN3ZGkuY28udWsiLCJuYmYiOjE2NzYyNTQxMTZ9.7jjL5I-BOhD-SC3NoaSNWNd1dG8jZXI8faRUK4YGocOlbMk4mzDl2AtJJ_0Wh9TJa-VAAmdPSzrRMjk7_cs0hw"

	teams1, err := session.ListAllTeams(token)
	if err != nil {
		fmt.Println(err)
	}

	for _, team := range teams1 {
		err = session.DeleteTeam(team.Id, token)
		if err != nil {
			fmt.Println(err)
		}
	}

	for _, team := range teams {
		addTeam, err := session.AddTeam(team, token)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(team)
		fmt.Println(addTeam)
		fmt.Println(addTeam.FileName)
	}
}

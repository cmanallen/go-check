package main

import (
	"os"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"log"
	"strconv"
	"time"

	"github.com/codegangsta/cli"
	"github.com/wsxiaoys/terminal"
)


type Server struct {
	Name	string
	URL		string
	Status	int
}

type ServerList struct {
	Servers []*Server
}

func new_server(name string, url string) *Server {
	return &Server{
		Name: name,
		URL: url,
	}
}

func status_check(url string) int {
	response, error := http.Get(url)
	if error != nil {
		return 0
	}
	return response.StatusCode
}

func output(servers []*Server) {
	for _, server := range servers {
		status := status_check(server.URL)
		terminal.Stdout.Color("y").Print(server.Name + " ").Color("c").Print("(" + server.URL + ")")
		if status == 200 || status == 301 {
			terminal.Stdout.Color("g").Print(" [", status, "]")
		} else if status == 400 || status == 404 || status == 500 {
			terminal.Stdout.Color("r").Print(" [", status, "]")
		} else {
			terminal.Stdout.Color("o").Print(" [", status, "]")
		}
		terminal.Stdout.Nl()
	}
	terminal.Stdout.Reset()
}

func (sl *ServerList) add_server_to_list(name string, url string) {
	sl.Servers = append(sl.Servers, new_server(name, url))
}

func (sl *ServerList) remove_server_from_list(name string) {
	servers := make([]*Server, 0)
	for _, server := range sl.Servers {
		if server.Name != name {
			servers = append(servers, server)
		}
	}
	sl.Servers = servers
}

func (sl *ServerList) load_file(fileName string) error {
	if _, error := os.Stat(fileName); os.IsNotExist(error) {
		os.Create(fileName)
	}
	
	file, error := os.Open(fileName)
	if error != nil {
		return error
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	error = decoder.Decode(sl)
	if error == io.EOF {
		return nil
	}
	return error
}

func (sl *ServerList) save_file(fileName string) error {
	data, error := json.MarshalIndent(sl, "", "    ")
	if error != nil {
		return error
	}
	error = ioutil.WriteFile(fileName, data, 0644)
	return error
}

func load(sl *ServerList, context *cli.Context) {
	timeout := time.Duration(context.GlobalInt("timeout"))
	if len(sl.Servers) == 0 {
		log.Fatalln("No servers found.")
	}
	output(sl.Servers)
}

func main() {
	app := cli.NewApp()
	app.Version = "0.1"
	app.Name = "server_check"
	app.Usage = "Quickly check your servers' status(es)."

	app.Commands = []cli.Command{
		{
			Name: "new",
			Usage: "Add new server to server list.",
			Action: func(c *cli.Context) {
				name := c.Args().First()
				url, err := strconv.Atoi(c.Args().Get(1))
				if err != nil {
					log.Fatalln(err)
				}

				config := &ServerList{}
				if err = config.load_file(c.GlobalString("config")); err != nil {
					log.Fatalln(err)
				}
				config.add_server_to_list(name, url)

				err = config.save_file(c.GlobalString("config"))
				if err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			Name: "remove",
			Usage: "Remove server from server list.",
			Action: func(c *cli.Context) {
				name := c.Args.First()
				config := &ServerList{}

				if error := config.load_file(c.GlobalString("config")); error != nil {
					log.Fatalln(error)
				}
				config.remove_server_from_list(name)
			},
		},
		{
			Name: "list",
			Usage: "Lists added servers.",
			Action: func(c *cli.Context) {
				config := &ServerList{}
				if err := config.load_file(c.GlobalString("config")); err != nil {
					log.Fatalln("Config file error: ", err)
				}

				for _, server := range config.Servers {
					terminal.Stdout.Color("y").Print(server.Name).Print(" ").Color("g").Print(server.URL).Nl().Reset()
				}
			},
		},
	}

	app.Action = func(c *cli.Context) {
		config := &ServerList{}
		if err := config.load_file(c.GlobalString("config")); err != nil {
			log.Fatalln("Config file error: ", err)
		}
		load(config, c)
	}

	app.Run(os.Args)
}
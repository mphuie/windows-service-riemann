package main

import (
	"log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/riemann/riemann-go-client"
	"context"
	"time"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"os"
)

type Config struct {
	Riemann_host string
	Host string
	States map[string]string
	Whitelist map[string]string
}

func main() {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	config := Config{}
	yaml_error := yaml.Unmarshal([]byte(data), &config)
	if yaml_error != nil {
		log.Fatal(yaml_error)
	}

	c := riemanngo.NewTcpClient(config.Riemann_host)
	err1 := c.Connect(5)
	if err1 != nil {
		panic(err1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()

	command := fmt.Sprintf("sc query state= all")
	cmd := exec.CommandContext(ctx, "cmd", "/C", command)
	out, err := cmd.Output()


	if ctx.Err() == context.DeadlineExceeded {
		hostname, _ := os.Hostname()
		riemanngo.SendEvent(c, &riemanngo.Event{
			Service: "service-collector",
			State: "critical",
			Metric: 0,
			Description: "Unable to gather service state!",
			Host: hostname,
			Ttl: 300,
		})
	}

	var serviceBlock = regexp.MustCompile(`(?s)SERVICE_NAME.*?\n\r\n`)
	var serviceState = regexp.MustCompile(`(?m)STATE\s+:\s\d+\s+(?P<state>.*?)$`)
	var serviceName = regexp.MustCompile(`(?m)SERVICE_NAME:\s+(.*?)$`)

	outString := string(out[:])
	matches := serviceBlock.FindAllString(outString, -1)

	events := []riemanngo.Event{}

	for i := 0; i< len(matches); i++ {

		state := serviceState.FindStringSubmatch(matches[i])[1]
		name := serviceName.FindStringSubmatch(matches[i])[1]
		ss := strings.TrimSpace(state)
		sn := strings.TrimSpace(name)

		if config.Whitelist[sn] != "" {

			var riemannState string
			if ss == config.Whitelist[sn] {
				riemannState = "ok"
			} else {
				riemannState = "critical"
			}

			events = append(events, riemanngo.Event{
				Service:     sn,
				State:       riemannState,
				Metric:      0,
				Description: fmt.Sprintf("%v: %v (desired %v)", sn, ss, config.Whitelist[sn]),
				Host:        config.Host,
				Ttl:         300,
			},
			)
			fmt.Printf("Sending status for %v\n", sn)
		} else {
			//fmt.Printf("Service %v not in whitelist!\n", sn)
		}

	}



	riemanngo.SendEvents(c, &events)



}
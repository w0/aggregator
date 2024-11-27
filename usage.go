package main

import (
	"fmt"
	"strings"
)

func usage(cmds map[string]func(*state, command) error) string {
	usage := strings.Builder{}
	usage.WriteString(fmt.Sprintln("usage: aggregator command <arguments>"))
	usage.WriteString(fmt.Sprintln("\tcommands:"))
	for k := range cmds {
		usage.WriteString(fmt.Sprintf("\t\t%s\n", k))
	}

	return usage.String()

}

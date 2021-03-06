package aini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/flynn/go-shlex"
)

type Hosts struct {
	input  *bufio.Reader
	Groups map[string][]Host
}

type Host struct {
	Name string
	Port int
}

func NewFile(f string) (*Hosts, error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return &Hosts{}, err
	}

	h, err := NewParser(bytes.NewReader(bs))
	if err != nil {
		return &Hosts{}, err
	}

	return h, nil
}

func NewParser(r io.Reader) (*Hosts, error) {
	input := bufio.NewReader(r)
	hosts := &Hosts{input: input}
	hosts.parse()
	return hosts, nil
}

func (h *Hosts) parse() error {
	scanner := bufio.NewScanner(h.input)
	activeGroupName := "ungrouped"
	h.Groups = make(map[string][]Host)
	h.Groups[activeGroupName] = make([]Host, 0)

	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")
		// fmt.Println(activeGroupName, ":", line)

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			replacer := strings.NewReplacer("[", "", "]", "")
			activeGroupName = replacer.Replace(line)

			if _, ok := h.Groups[activeGroupName]; !ok {
				h.Groups[activeGroupName] = make([]Host, 0)
			}
		} else if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || line == "" {
			// do nothing
		} else if activeGroupName != "" {
			parts, err := shlex.Split(line)
			if err != nil {
				fmt.Println("couldn't tokenizer ", line)
			}
			host := getHost(parts)
			h.Groups[activeGroupName] = append(h.Groups[activeGroupName], host)
		}
	}
	return nil
}

func (h *Hosts) Match(m string) []Host {
	matchedHosts := make([]Host, 0, 5)
	for _, hosts := range h.Groups {
		for _, host := range hosts {
			if m, err := path.Match(m, host.Name); err == nil && m {
				matchedHosts = append(matchedHosts, host)
			}
		}
	}
	return matchedHosts
}

func getHost(parts []string) Host {
	hostname := parts[0]
	port := 22
	if (strings.Contains(hostname, "[") &&
		strings.Contains(hostname, "]") &&
		strings.Contains(hostname, ":") &&
		(strings.LastIndex(hostname, "]") < strings.LastIndex(hostname, ":"))) ||
		(!strings.Contains(hostname, "]") && strings.Contains(hostname, ":")) {

		splithost := strings.Split(hostname, ":")
		if i, err := strconv.Atoi(splithost[1]); err == nil {
			port = i
		}
		hostname = splithost[0]
	}
	host := Host{Name: hostname, Port: port}

	return host
}

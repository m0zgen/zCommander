package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

// Config - structure for storing server configuration
type Config struct {
	Groups          map[string][]string `yaml:"groups"`
	CreationTimeout time.Duration       `yaml:"creation_timeout"`
	UsersFile       string              `yaml:"users_file"`
	AddCommand      string              `yaml:"add_command"`
	RemoveCommand   string              `yaml:"remove_command"`
}

// readFile - read content of file and return it as slice of strings
func readFile(filePath string) []string {

	// Open file for reading
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer file.Close()

	// Scan file line by line
	scanner := bufio.NewScanner(file)

	// Fill slice with words
	var words []string
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	// Check for errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return nil
	}

	return words
}

// sendCommandToServers - send command to servers
func sendCommandToServers(servers []string, command string, timeout time.Duration) {
	var wg sync.WaitGroup
	for _, server := range servers {
		wg.Add(1)
		go func(serverURL string) {
			defer wg.Done()

			// Ignore TLS certificate errors (work with self-signed certificates)
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}

			url := serverURL + command
			resp, err := client.Get(url)
			if err != nil {
				fmt.Printf("Error sending command to server %s: %v\n", serverURL, err)
				return
			}
			defer resp.Body.Close()
			fmt.Printf("Response from server %s: %s\n", serverURL, resp.Status)

			// Read response from server
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response body from server %s: %v\n", serverURL, err)
				return
			}

			// Show response from server
			fmt.Printf("Response from server %s: %s\n", serverURL, string(body))
		}(server)
		time.Sleep(timeout)
	}
	wg.Wait()
}

func main() {

	// Read command line arguments
	// TODO: add - `get_command`, `download_command`, `upload_command`. Add flags for `get`, `download`, `upload`
	configFile := flag.String("config", "", "path to config file")
	groupName := flag.String("group", "", "name of server group")
	command := flag.String("command", "", "command to send to servers")
	creationTimeout := flag.Duration("creation_timeout", 0, "interval between command creation")
	usersFile := flag.String("file", "", "path to file with users list")
	massFlag := flag.Bool("mass", false, "mass command to all servers (create users from file)")
	addFlag := flag.Bool("add", false, "add user to all servers")
	removeFlag := flag.Bool("remove", false, "remove user from all servers")
	flag.Parse()

	// Check required arguments
	if *configFile == "" || *groupName == "" {
		fmt.Println("Usage: go run main.go -config=config.yaml -group=testlocal1 -command='/add?user=user_u-2foo'")
		return
	}

	// Read configuration file
	data, err := os.ReadFile(*configFile)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return
	}

	// Decode YAML configuration
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Error decoding config file: %v\n", err)
		return
	}

	// Check if group exists in config file
	servers, ok := config.Groups[*groupName]
	if !ok {
		fmt.Printf("Group %s not found in config file\n", *groupName)
		return
	}

	// Check if creation_timeout is set and set it to default value from config if not
	if *creationTimeout == 0 {
		if config.CreationTimeout == 0 {
			*creationTimeout = 10 * time.Second // Значение по умолчанию
		} else {
			*creationTimeout = config.CreationTimeout * time.Second
		}
	}

	// Manage multiple users
	// Check if mass flag is set and send command to servers
	if *massFlag {

		// If add or remove flags are set, then we should use the users file
		if *addFlag || *removeFlag {

			// Check if file argument is set and set it to default value from config if not
			if *usersFile == "" {
				*usersFile = config.UsersFile
			}
		}

		// Read file with users
		users := readFile(*usersFile)

		for _, user := range users {

			if *addFlag {
				*command = config.AddCommand + user
			} else if *removeFlag {
				*command = config.RemoveCommand + user
			}

			fmt.Printf("User: %s. Command: %s\n", user, *command)

			// Send command to servers
			sendCommandToServers(servers, *command, *creationTimeout)
		}

		return

	}

	if *command == "" {
		fmt.Println("Command is not specified")
		return
	}

	// Manage single user. Send command to servers
	sendCommandToServers(servers, *command, *creationTimeout)
}

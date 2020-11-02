package main

import (
	"bufio"
	"flag"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

type Config struct {
	hostname     string
	user         string
	identityfile string
	source       string
	target       string
}

func loadConfig() Config {
	sourceDir, _ := os.Getwd()
	sourcePath := strings.Split(sourceDir, "/")
	targetDir := sourcePath[len(sourcePath) - 1]

	m := merge(
		map[string]string{
			"source": sourceDir,
			"target": targetDir,
		},
		loadConfigFile(path.Join(os.Getenv("HOME"), ".config/gobin")),
		loadConfigFile(path.Join(os.Getenv("PWD"), ".gobin")),
		loadEnvs(),
		loadArgs(),
	)
	m = merge(m, loadSshConfigFor(m["hostname"]))

	config := Config{
		m["hostname"],
		m["user"],
		m["identityfile"],
		m["source"],
		m["target"],
	}

	if config.hostname == "" {
		log.Fatal("Missing required parameter: hostname")
	}
	if config.user == "" {
		log.Fatal("Missing required parameter: user")
	}
	if config.identityfile == "" {
		log.Fatal("Missing required parameter: identityfile")
	}
	if strings.HasPrefix(config.identityfile, "~/") {
		config.identityfile = path.Join(os.Getenv("HOME"), m["identityfile"][1:])
	}
	if _, err := os.Stat(config.identityfile); os.IsNotExist(err) {
		log.Fatalf("can't find identity file: %s", config.identityfile)
	}

	return config
}

func loadSshConfigFor(remote string) map[string]string {
	file, err := os.Open(path.Join(os.Getenv("HOME"), ".ssh/config"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		log.Fatalf("can't open [~/.ssh/config]: %s", err.Error())
	}
	defer file.Close()

	hostPat := regexp.MustCompile("^host\\s+(\\S+)")
	kvPat := regexp.MustCompile("^\\s+(\\S+)\\s+(\\S+)")

	var config map[string]string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())

		if matches := hostPat.FindStringSubmatch(line); matches != nil {
			host := matches[1]
			if host == remote {
				config = map[string]string{}
				continue
			}
			if host != "" && config != nil {
				return config
			}
			continue
		}

		if config == nil {
			continue
		}

		if matches := kvPat.FindStringSubmatch(line); matches != nil {
			k := matches[1]
			v := matches[2]
			if k == "hostname" || k == "user" || k == "identityfile" {
				config[k] = v
			}
		}
	}

	return config
}

func loadConfigFile(path string) map[string]string {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		log.Fatalf("can't open [%s]: %s", path, err.Error())
	}
	config, err := godotenv.Parse(file)
	if err != nil {
		log.Fatalf("can't parse [%s]: %s", path, err.Error())
	}
	file.Close()
	return config
}

func loadEnvs() map[string]string {
	result := map[string]string{}
	var v string
	v = os.Getenv("GOBIN_HOSTNAME")
	if v != "" {
		result["hostname"] = v
	}
	v = os.Getenv("GOBIN_USER")
	if v != "" {
		result["user"] = v
	}
	v = os.Getenv("GOBIN_IDENTITYFILE")
	if v != "" {
		result["identityfile"] = v
	}
	v = os.Getenv("GOBIN_SOURCE")
	if v != "" {
		result["source"] = v
	}
	v = os.Getenv("GOBIN_TARGET")
	if v != "" {
		result["target"] = v
	}
	return result
}

func loadArgs() map[string]string {
	result := map[string]string{}
	hostname := flag.String("h", "", "Remote hostname")
	user := flag.String("u", "", "Remote user")
	identityfile := flag.String("i", "", "SSH identity file")
	source := flag.String("s", "", "Source directory")
	target := flag.String("t", "", "Target directory")
	flag.Parse()
	if *hostname != "" {
		result["hostname"] = *hostname
	}
	if *user != "" {
		result["user"] = *user
	}
	if *identityfile != "" {
		result["identityfile"] = *identityfile
	}
	if *source != "" {
		result["source"] = *source
	}
	if *target != "" {
		result["target"] = *target
	}
	return result
}

func merge(maps ...map[string]string) map[string]string {
	result := map[string]string{}
	for _, m := range maps {
		if m != nil {
			for k, v := range m {
				result[k] = v
			}
		}
	}
	return result
}

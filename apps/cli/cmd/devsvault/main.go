package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const defaultAPI = "http://localhost:8080"

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printHelp(stdout)
		return nil
	}
	switch args[0] {
	case "login":
		return login(args[1:], stdout)
	case "pull":
		return pull(args[1:], stdout)
	case "inject":
		return inject(args[1:], stdout)
	case "run":
		return runWithSecrets(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func login(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	apiURL := fs.String("api", getenv("DEVSVAULT_API_URL", defaultAPI), "API URL")
	subject := fs.String("subject", "admin@example.local", "OIDC subject for development login")
	actorType := fs.String("type", "user", "actor type: user or service")
	if err := fs.Parse(args); err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]string{"subject": *subject, "actor_type": *actorType})
	resp, err := http.Post(strings.TrimRight(*apiURL, "/")+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d", resp.StatusCode)
	}
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return err
	}
	if payload.AccessToken == "" {
		return errors.New("login response did not include token")
	}
	if err := saveConfig(config{APIURL: *apiURL, Token: payload.AccessToken}); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "login ok")
	return nil
}

func pull(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("pull", flag.ContinueOnError)
	path := fs.String("path", "", "logical path workspace/project/env/name")
	format := fs.String("format", "env", "output format: env or json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	resolved, err := resolve(*path)
	if err != nil {
		return err
	}
	name := secretName(resolved.Path)
	if *format == "json" {
		return json.NewEncoder(stdout).Encode(map[string]string{name: resolved.Value})
	}
	fmt.Fprintf(stdout, "%s=%s\n", name, shellQuote(resolved.Value))
	return nil
}

func inject(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("inject", flag.ContinueOnError)
	path := fs.String("path", "", "logical path workspace/project/env/name")
	name := fs.String("name", "", "environment variable name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	resolved, err := resolve(*path)
	if err != nil {
		return err
	}
	varName := *name
	if varName == "" {
		varName = secretName(resolved.Path)
	}
	fmt.Fprintf(stdout, "%s=%s\n", varName, shellQuote(resolved.Value))
	return nil
}

func runWithSecrets(args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	path := fs.String("path", "", "logical path workspace/project/env/name")
	name := fs.String("name", "", "environment variable name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	command := fs.Args()
	if len(command) == 0 {
		return errors.New("run requires a command after flags")
	}
	resolved, err := resolve(*path)
	if err != nil {
		return err
	}
	varName := *name
	if varName == "" {
		varName = secretName(resolved.Path)
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), varName+"="+resolved.Value)
	return cmd.Run()
}

type config struct {
	APIURL string `json:"api_url"`
	Token  string `json:"token"`
}

type resolvedSecret struct {
	Path    string `json:"path"`
	Version int    `json:"version"`
	Value   string `json:"value"`
}

func resolve(path string) (resolvedSecret, error) {
	if path == "" {
		return resolvedSecret{}, errors.New("--path is required")
	}
	cfg, err := loadConfig()
	if err != nil {
		return resolvedSecret{}, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(cfg.APIURL, "/")+"/api/v1/secrets/resolve?path="+path, nil)
	if err != nil {
		return resolvedSecret{}, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)
	resp, err := client.Do(req)
	if err != nil {
		return resolvedSecret{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resolvedSecret{}, fmt.Errorf("secret resolve failed with status %d", resp.StatusCode)
	}
	var payload resolvedSecret
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return resolvedSecret{}, err
	}
	return payload, nil
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "devsvault", "config.json"), nil
}

func saveConfig(cfg config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func loadConfig() (config, error) {
	path, err := configPath()
	if err != nil {
		return config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return config{}, errors.New("not logged in; run devsvault login")
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return config{}, err
	}
	if cfg.Token == "" || cfg.APIURL == "" {
		return config{}, errors.New("invalid CLI config; run devsvault login")
	}
	return cfg, nil
}

func secretName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return "SECRET"
	}
	return strings.ToUpper(strings.ReplaceAll(parts[len(parts)-1], "-", "_"))
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func printHelp(stdout io.Writer) {
	fmt.Fprintln(stdout, "Devsvault CLI")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "Commands:")
	fmt.Fprintln(stdout, "  login  --api http://localhost:8080 --subject admin@example.local")
	fmt.Fprintln(stdout, "  pull   --path workspace/project/dev/DATABASE_URL")
	fmt.Fprintln(stdout, "  inject --path workspace/project/dev/API_TOKEN --name API_TOKEN")
	fmt.Fprintln(stdout, "  run    --path workspace/project/dev/API_TOKEN --name API_TOKEN -- npm start")
}

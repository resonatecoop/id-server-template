package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/oxtoacart/bpool"
)

var (
	templates map[string]*template.Template
	bufpool   *bpool.BufferPool
	loaded    = false
)

// load and parse hash.json from data dir
func loadData(dirname string) (map[string]interface{}, error) {
	var result map[string]interface{}

	file, err := os.Open(filepath.Join("data", dirname, "hash.json"))

	if err != nil {
		return nil, err
	}

	byteValue, err := ioutil.ReadAll(file)

	if err != nil {
		file.Close()
		return nil, err
	}

	err = file.Close()

	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(byteValue), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// renderTemplate is a wrapper around template.ExecuteTemplate.
// It writes into a bytes.Buffer before writing to the http.ResponseWriter to catch
// any errors resulting from populating the template.
func renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	loadTemplates()

	// Ensure the template exists in the map.
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("The template %s does not exist", name)
	}

	js, err := loadData("js")

	if err != nil {
		return err
	}

	data["javascript"] = js["main.js"]

	style, err := loadData("css")

	if err != nil {
		return err
	}

	data["stylesheet"] = style["index.css"]

	// Create a buffer to temporarily write to and check if any errors were encountered.
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err = tmpl.ExecuteTemplate(buf, "base", data)
	if err != nil {
		return err
	}

	// The X-Frame-Options HTTP response header can be used to indicate whether
	// or not a browser should be allowed to render a page in a <frame>,
	// <iframe> or <object> . Sites can use this to avoid clickjacking attacks,
	// by ensuring that their content is not embedded into other sites.
	w.Header().Set("X-Frame-Options", "deny")
	// Set the header and write the buffer to the http.ResponseWriter
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = buf.WriteTo(w)
	if err != nil {
		return err
	}
	return nil
}

func loadTemplates() {
	if loaded {
		return
	}

	templates = make(map[string]*template.Template)

	bufpool = bpool.NewBufferPool(64)

	layoutTemplates := map[string][]string{
		"web/layouts/outside.html": {
			"./web/includes/join.html",
			"./web/includes/login.html",
			"./web/includes/password_reset.html",
			"./web/includes/password_reset_update_password.html",
			"./web/includes/home.html",
		},
		"web/layouts/inside.html": {
			"./web/includes/authorize.html",
			"./web/includes/client.html",
			"./web/includes/account.html",
			"./web/includes/account_settings.html",
			"./web/includes/membership.html",
			"./web/includes/profile.html",
			"./web/includes/checkout.html",
		},
	}

	for layout, includes := range layoutTemplates {
		for _, include := range includes {
			files := []string{include, layout}
			templates[filepath.Base(include)] = template.Must(template.ParseFiles(files...))
		}
	}

	loaded = true
}

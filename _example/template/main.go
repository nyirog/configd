// Package main demonstrate the use of configd with a simple config.
// ./config.tmpl is rendered from Config during Update.
package main

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/nyirog/configd"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
)

func parseTemplate(templateFile string) *template.Template {
	baseName := path.Base(templateFile)
	return template.Must(template.New(baseName).ParseFiles(templateFile))
}

type Config struct {
	SomeInt    int32
	SomeString string
}

func (config *Config) Update(change *prop.Change) *dbus.Error {
	configd.SetStructField(config, change)

	err := renderer.Execute(os.Stdout, config)
	if err != nil {
		return dbusError("org.freedesktop.DBus.Properties.Error.RenderingError", err)
	}

	return nil
}

func (config *Config) CreatePropertyMap() map[string]interface{} {
	return configd.CreateStructMap(config)
}

func dbusError(name string, err error) *dbus.Error {
	body := []interface{}{err.Error()}
	return dbus.NewError(name, body)
}

var renderer *template.Template

func main() {
	renderer = parseTemplate("./config.tpl")

	config := Config{SomeInt: 42, SomeString: "egg"}

	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	reply, err := conn.RequestName("org.configd",
		dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Fprintln(os.Stderr, "name already taken")
		os.Exit(1)
	}

	configd.CreateNode(conn, "/config", "org.configd.Config", &config)

	fmt.Println("Listening on org.configd /config ...")

	select {}
}

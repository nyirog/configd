// Package main demonstrate the use of configd with a simple config. The Config
// is validated during Update.
package main

import (
	"fmt"
	"os"

	"github.com/nyirog/configd"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"

	"gopkg.in/go-playground/validator.v9"
)

type Config struct {
	SomeInt    int32  `validate:"max=50"`
	SomeString string `validate:"len=3"`
}

func (config *Config) Update(change *prop.Change) *dbus.Error {
	configd.SetStructField(config, change)

	err := validate.Struct(config)
	if err != nil {
		return validationError(err.(validator.ValidationErrors))
	}

	fmt.Println("Update", change.Name, change.Value)
	return nil
}

func (config *Config) CreatePropertyMap() map[string]interface{} {
	return configd.CreateStructMap(config)
}

func validationError(err validator.ValidationErrors) *dbus.Error {
	body := make([]interface{}, len(err))
	for i, e := range err {
		body[i] = fmt.Sprintf(
			"%s[%s=%s] <> %v", e.Field(), e.Tag(), e.Param(), e.Value(),
		)
	}
    return dbus.NewError("org.freedesktop.DBus.Properties.Error.ValidationErrors", body)
}

var validate *validator.Validate

func main() {
	validate = validator.New()

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

// Package configd helps to expose an application config to dbus.
package configd

import (
	"reflect"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

// UpdateError is a general error during the Update call of
// org.freedesktop.DBus.Properties.Set.
var UpdateError = dbus.NewError("org.freedesktop.DBus.Properties.Error.UpdateError", nil)

// Updatable is the required interface which has to be implemented by the
// config struct to able to be exposed to dbus.
type Updatable interface {
	Update(*prop.Change) *dbus.Error

	CreatePropertyMap() map[string]interface{}
}

func createProps(config Updatable) map[string]*prop.Prop {
	var props = make(map[string]*prop.Prop)

	for name, value := range config.CreatePropertyMap() {
		props[name] = &prop.Prop{
			value,
			true,
			prop.EmitTrue,
			config.Update,
		}
	}

	return props
}

// CreateStructMap is a helper function to implement the CreatePropertyMap
// method of the Updatable interface. CreateStructMap will extraxt the public
// field from the given config.
func CreateStructMap(config interface{}) map[string]interface{} {
	var struct_map = make(map[string]interface{})

	config_value := reflect.ValueOf(config).Elem()
	config_type := config_value.Type()

	for i := 0; i < config_type.NumField(); i++ {
		field_type := config_type.Field(i)
		if strings.Title(field_type.Name) != field_type.Name {
			continue
		}

		struct_map[field_type.Name] = config_value.Field(i).Interface()
	}

	return struct_map
}

// SetStructField is a helper function which can be used in the Update method
// of the Updatable interface. SetStructField updates the Change.Name field of
// the config. SetStructField should be the first call of the Update
// implementation of the config.
func SetStructField(config interface{}, change *prop.Change) {
	config_value := reflect.ValueOf(config).Elem()
	config_value.FieldByName(change.Name).Set(reflect.ValueOf(change.Value))
}

// CreateNode export the config under the iface name of the path object to dbus.
func CreateNode(conn *dbus.Conn, path dbus.ObjectPath, iface string, config Updatable) {
	propsSpec := map[string]map[string]*prop.Prop{
		iface: createProps(config),
	}
	props := prop.New(conn, path, propsSpec)
	node := &introspect.Node{
		Name: string(path),
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			{
				Name:       iface,
				Properties: props.Introspection(iface),
			},
		},
	}
	conn.Export(introspect.NewIntrospectable(node), path,
		"org.freedesktop.DBus.Introspectable")
}

package config

import (
	"errors"
	"flag"
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

const (
	overwriteEnvKey = "overwrite-env"
	overwriteArgKey = "overwrite-arg"
)

func ReadConfiguration(configPath string, c interface{}) error {
	t := reflect.TypeOf(c)

	switch t.Kind() {
	case reflect.Ptr, reflect.Interface:
		viper.SetConfigName(configPath)
		viper.AddConfigPath(".")

		err := viper.ReadInConfig()
		if err != nil {
			return fmt.Errorf("error reading config file, %s", err)
		}

		val := t.Elem()
		fields := val.NumField()
		for i := 0; i < fields; i++ {
			child := val.Field(i)
			if err := overwriteFields("", child, viper.GetViper()); err != nil {
				return err
			}
		}

		flag.Parse()

		err = viper.Unmarshal(c)
		if err != nil {
			return err
		}
		return nil
	case reflect.Struct:
		return errors.New("configuration to load need to be a pointer")
	default:
		return errors.New("invalid configuration structure")

	}
}

func overwriteFields(parent string, f reflect.StructField, v *viper.Viper) error {
	prefix := ""
	if len(parent) > 0 {
		prefix = parent + "."
	}

	switch f.Type.Kind() {
	case reflect.Struct:
		t := f.Type
		fields := t.NumField()
		for i := 0; i < fields; i++ {
			child := t.Field(i)
			if err := overwriteFields(prefix+f.Name, child, v); err != nil {
				return err
			}
		}
	case reflect.Int:
		if err := overwriteValue(prefix, f, v, setIntFlag); err != nil {
			return err
		}
	case reflect.String:
		if err := overwriteValue(prefix, f, v, setStringFlag); err != nil {
			return err
		}
	case reflect.Bool:
		if err := overwriteValue(prefix, f, v, setBoolFlag); err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid field[%s] type[%s]", f.Name, f.Type.Name())
	}
	return nil
}

func overwriteValue(prefix string, f reflect.StructField, v *viper.Viper, setFlag func(v *viper.Viper, flagVal string, key string)) error {
	tag := f.Tag
	if len(tag) > 0 {
		val, ok := tag.Lookup(overwriteEnvKey)
		if ok {
			err := v.BindEnv(prefix+f.Name, val)
			if err != nil {
				return err
			}
		}
		val, ok = tag.Lookup(overwriteArgKey)
		if ok {
			key := prefix + f.Name
			setFlag(v, val, key)
		}
	}
	return nil
}

func setStringFlag(v *viper.Viper, flagVal string, key string) {
	viper.Set(key, flag.String(flagVal, viper.GetString(key), ""))
}

func setBoolFlag(v *viper.Viper, flagVal string, key string) {
	viper.Set(key, flag.Bool(flagVal, viper.GetBool(key), ""))
}

func setIntFlag(v *viper.Viper, flagVal string, key string) {
	viper.Set(key, flag.Int(flagVal, viper.GetInt(key), ""))
}

/*
 *  grml - A simple build automation tool written in Go
 *  Copyright (C) 2017  Roland Singer <roland.singer[at]desertbit.com>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package manifest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/desertbit/grml/internal/options"
	"gopkg.in/yaml.v2"
)

const (
	Version = 2
)

type Manifest struct {
	Version int    `yaml:"version"`
	Project string `yaml:"project"`

	EnvFiles    []string               `yaml:"envs"`
	Env         yaml.MapSlice          `yaml:"env"` // Use MapSlice to preserve order.
	Options     map[string]interface{} `yaml:"options"`
	Interpreter string                 `yaml:"interpreter"`
	Import      []string               `yaml:"import"`
	Commands    Commands               `yaml:"commands"`
}

type Commands map[string]*Command

type Command struct {
	Alias    []string `yaml:"alias"`
	Help     string   `yaml:"help"`
	Args     []string `yaml:"args"`
	Deps     []string `yaml:"deps"`
	Exec     string   `yaml:"exec"`
	Include  string   `yaml:"include"`
	Commands Commands `yaml:"commands"`
}

func (cs Commands) Count() (n int) {
	n = len(cs)
	for _, c := range cs {
		n += c.Commands.Count()
	}
	return
}

func (m *Manifest) EvalEnv(parentEnv map[string]string) (env map[string]string, err error) {
	// Define a function which evaluates environment variables and
	// appends them to the given map.
	appendEnvVars := func(ymap yaml.MapSlice, vars map[string]string) map[string]string {
		for _, i := range ymap {
			key := fmt.Sprintf("%v", i.Key)
			value := fmt.Sprintf("%v", i.Value)

			// Replace already existing variable names with their corresponding values.
			for k, v := range vars {
				value = strings.Replace(value, fmt.Sprintf("${%s}", k), v, -1)
			}
			for k, v := range parentEnv {
				value = strings.Replace(value, fmt.Sprintf("${%s}", k), v, -1)
			}
			vars[key] = value
		}
		return vars
	}

	// Read environment variable files if applicable.
	env = make(map[string]string)
	for _, ef := range m.EnvFiles {
		var content []byte
		content, err = os.ReadFile(ef)
		if err != nil {
			err = fmt.Errorf("unable to read env file '%s': %v", ef, err)
			return
		}

		// Unmarshal environment variable file in an order preserved manner.
		var ordered yaml.MapSlice
		err = yaml.Unmarshal(content, &ordered)
		if err != nil {
			err = fmt.Errorf("unable to unmarshal env data of file '%s': %v", ef, err)
			return
		}

		// Prepare and evaluate the environment variables.
		env = appendEnvVars(ordered, env)
	}

	// Prepare and evaluate the environment variables.
	env = appendEnvVars(m.Env, env)

	// Merge missing values from the parent environment.
	for k, v := range parentEnv {
		if _, ok := env[k]; !ok {
			env[k] = v
		}
	}
	return
}

func (m *Manifest) ParseOptions() (o *options.Options, err error) {
	o = options.New()
	for name, i := range m.Options {
		switch v := i.(type) {
		case bool:
			if _, ok := o.Bools[name]; ok {
				err = fmt.Errorf("duplicate option: %v", name)
				return
			}
			o.Bools[name] = v

		case []interface{}:
			if _, ok := o.Choices[name]; ok {
				err = fmt.Errorf("duplicate option: %v", name)
				return
			} else if len(v) == 0 {
				err = fmt.Errorf("invalid option: %v", name)
				return
			}

			list := make([]string, len(v))
			for i, iv := range v {
				list[i] = fmt.Sprintf("%v", iv)
			}

			o.Choices[name] = &options.Choice{
				Active:  list[0],
				Options: list,
			}

		default:
			err = fmt.Errorf("invalid option: %v: %v", name, i)
			return
		}
	}
	return
}

// Parse a grml build file.
func Parse(path string) (m *Manifest, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	m = &Manifest{}
	err = yaml.UnmarshalStrict(data, m)
	if err != nil {
		return
	}

	// Validate.
	if m.Version != Version {
		err = fmt.Errorf("incompatible grml version: file=%v current=%v", m.Version, Version)
		return
	} else if m.Project == "" {
		err = fmt.Errorf("no project name set")
		return
	}

	// Parse inlcudes.
	rootPath := filepath.Dir(path)
	err = parseIncludes(rootPath, m.Commands)
	if err != nil {
		return
	}

	return
}

func parseIncludes(rootPath string, cmds Commands) (err error) {
	for _, cmd := range cmds {
		if cmd.Include == "" {
			continue
		}

		var data []byte
		data, err = ioutil.ReadFile(filepath.Join(rootPath, cmd.Include))
		if err != nil {
			return
		}

		err = yaml.UnmarshalStrict(data, cmd)
		if err != nil {
			return
		}

		err = parseIncludes(rootPath, cmd.Commands)
		if err != nil {
			return
		}
	}
	return
}

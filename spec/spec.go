/*
 *  Grumble - A simple build automation tool written in Go
 *  Copyright (C) 2016  Roland Singer <roland.singer[at]desertbit.com>
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

package spec

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

//############//
//### Spec ###//
//############//

// Spec defines a grumble build file.
type Spec struct {
	//Options map[string]interface{} TODO
	Env     map[string]string
	Targets map[string]*Target
}

// EnvToSlice maps the environment variables to an os.exec Env slice.
func (s Spec) EnvToSlice() (env []string) {
	env = make([]string, len(s.Env))
	i := 0

	for k, v := range s.Env {
		env[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	return
}

// DefaultTarget returns the default run target if specified.
// Otherwise nil is returned.
func (s Spec) DefaultTarget() *Target {
	for _, t := range s.Targets {
		if t.Default {
			return t
		}
	}
	return nil
}

//######################//
//### Spec - Private ###//
//######################//

func (s *Spec) evaluateVars(str string, env map[string]string) string {
	if env != nil {
		for key, value := range env {
			key = fmt.Sprintf("${%s}", key)
			str = strings.Replace(str, key, value, -1)
		}
	}

	for key, value := range s.Env {
		key = fmt.Sprintf("${%s}", key)
		str = strings.Replace(str, key, value, -1)
	}

	return str
}

// targetWithOutput returns the target which creates the given output.
func (s Spec) targetWithOutput(o string) *Target {
	o = filepath.Clean(o)
	for _, t := range s.Targets {
		for _, to := range t.Outputs {
			if filepath.Clean(to) == o {
				return t
			}
		}
	}
	return nil
}

//##############//
//### Public ###//
//##############//

// ParseSpec parses a grumble build file.
func ParseSpec(path string, env map[string]string) (s *Spec, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	s = new(Spec)
	err = yaml.Unmarshal(data, s)
	if err != nil {
		return
	}

	// Evaluate the environment variables.
	for key, value := range s.Env {
		s.Env[key] = s.evaluateVars(value, env)
	}

	// Initialize the private target values.
	for name, t := range s.Targets {
		err = t.init(name, s)
		if err != nil {
			err = fmt.Errorf("target '%s': %v", name, err)
			return
		}
	}

	return
}

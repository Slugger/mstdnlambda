package devenv

/*
	mstdnlambda
	Copyright (C) 2022 Battams, Derek <derek@battams.ca>

	This program is free software; you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation; either version 2 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License along
	with this program; if not, write to the Free Software Foundation, Inc.,
	51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type ErrDevEnv error
type errDevEnv struct{}

func (e *errDevEnv) Error() string {
	return "devenv error"
}

var devEnvError = &errDevEnv{}

type envMap map[string]string

func (e *envMap) String() string {
	vals := make([]string, len(*e))
	i := 0
	for k, v := range *e {
		vals[i] = fmt.Sprintf("[%s=%s]", k, v)
		i++
	}
	return strings.Join(vals, ",")
}

func (e *envMap) Set(val string) error {
	pair := strings.SplitN(val, "=", 2)
	if len(pair) != 2 {
		return fmt.Errorf("%w: invalid env var '%s'; expected var=val", devEnvError, val)
	}
	(*e)[pair[0]] = pair[1]
	return nil
}

var isDevEnv bool
var eventFile string
var env envMap

func init() {
	env = make(envMap)

	flag.StringVar(&eventFile, "event", "event.json", "SQS event input file")
	flag.BoolVar(&isDevEnv, "devenv", false, "must be set to true to trigger dev mode")
	flag.Var(&env, "env", "specify an environment variable to set for the lambda execution; multiple -env can be specified; format is var=val")
}

func IsActive() bool {
	return isDevEnv
}

func GetEventData() ([]byte, error) {
	bytes, err := os.ReadFile(eventFile)
	if err != nil {
		return nil, fmt.Errorf("error reading event file: %w", err)
	}
	return bytes, nil
}

func InitArgs() {
	if !IsActive() {
		return
	}

	flag.Visit(func(f *flag.Flag) {
		if f.Name != "env" {
			return
		}
		for k, v := range env {
			if err := os.Setenv(k, v); err != nil {
				panic(err)
			}
		}
	})
}

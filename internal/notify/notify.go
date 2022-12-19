package notify

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
	"github.com/sirupsen/logrus"
	"github.com/slugger/mstdnlambda/internal/devenv"
	"github.com/slugger/mstdnlambda/internal/logging"
)

// Notifier defines the contract for receivers of incoming push notifications
type Notifier interface {
	// Send delivers the given message string to this Notifier; returns nil on success or non-nil in case of an error
	Send(message string) error
}

// New returns a default implementation of Notifier
func New(target string) Notifier {
	if devenv.IsActive() {
		return &devNotifier{
			target: target,
		}
	}
	return newSns(target)
}

type devNotifier struct {
	target string
}

var devLog *logrus.Entry

func init() {
	devLog = logging.GetLogForCategory(logging.DevEnvCategory)
}

func (n *devNotifier) Send(message string) error {
	devLog.WithField("target", n.target).Info(message)
	return nil
}

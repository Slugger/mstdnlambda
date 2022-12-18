package notify

import (
	"github.com/sirupsen/logrus"
	"github.com/slugger/mstdnlambda/internal/devenv"
	"github.com/slugger/mstdnlambda/internal/logging"
)

type Notifier interface {
	Send(message string) error
}

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

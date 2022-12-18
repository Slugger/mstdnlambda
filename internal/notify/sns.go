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
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/sirupsen/logrus"
	"github.com/slugger/mstdnlambda/internal/cfg"
	"github.com/slugger/mstdnlambda/internal/logging"
)

type snsNotifier struct {
	topicArn string
}

var snsLog *logrus.Entry
var snsSession *session.Session

func newSns(topicArn string) Notifier {
	if snsLog == nil {
		snsLog = logging.GetLogForCategory(logging.SnsNotificationCategory)
		snsLog.Debug("sns log initialized")
	}
	if snsSession == nil {
		snsSession = session.Must(session.NewSessionWithOptions(session.Options{
			Config: *aws.NewConfig().WithRegion(cfg.Cfg.AwsRegion()),
		}))
		snsLog.WithField("awsregion", cfg.Cfg.AwsRegion()).Debugf("sns session initialized")
	}
	return &snsNotifier{
		topicArn: topicArn,
	}
}

func (n *snsNotifier) Send(message string) error {
	log := snsLog.WithField("target", n.topicArn)
	svc := sns.New(snsSession)
	req := sns.PublishInput{
		TopicArn: &n.topicArn,
		Message:  &message,
	}
	resp, err := svc.Publish(&req)
	if err == nil {
		log.WithField("response", resp.String()).Debug("sns delivered")
	} else {
		err = fmt.Errorf("[sns publish failed] %w", err)
	}
	return err
}

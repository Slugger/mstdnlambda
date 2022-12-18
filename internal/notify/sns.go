package notify

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

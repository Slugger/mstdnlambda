package main

/*
	The handler command is an AWS Lambda compatible command that will
	process Mastodon push notifications via AWS LambdaFunctionURLRequest
	events received from a Mastodon instance and publish them to one or
	more SNS targets.

	The SNS targets are encoded into the request path. The lambda must be
	configured such that it can publish notifications to _ALL_ of the
	targets that are encoded in the request path.  If any target fails to
	publish then the request will return a failure to the caller, which
	usually results in retries and duplicate messages being received.
	Because of this, subscribers to the target topics _MUST_ be idempodent
	and must be prepared to process the same message multiple times.

	The lambda's function URL must be requested with a path such that each
	segment of the path is a URL safe base64 encoding of an SNS topic ARN.
	Each "directory" in the request URL is considered an encoded topic ARN
	and the notification received from Mastodon will be published to each
	ARN found in the request path.

	Example:
	https://abcxyz1234.lambda-url.us-east-1.on.aws/arnEncoding1/arnEncoding2/arnEncoding3

	A push notification from Mastodon sent to the above url implies that
	there are three target SNS topics for this request. Each of the arnEncodingX
	segments found in the path are decoded and used as targets.

	A full description of how to setup a lambda based push notification receiver
	is available at the project site:

	https://gitlab.com/ddb_db/mstdnlambda

	This command is also capable of running in development mode, which reads an event
	from a json file on the local filesystem and "publishes" the events to a log file
	intead of interacting with AWS.  See the project site for more details.
*/

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	"github.com/slugger/mstdnlambda/internal/cfg"
	"github.com/slugger/mstdnlambda/internal/devenv"
	"github.com/slugger/mstdnlambda/internal/http"
	"github.com/slugger/mstdnlambda/internal/jwt"
	"github.com/slugger/mstdnlambda/internal/logging"
	"github.com/slugger/mstdnlambda/internal/notify"
	"github.com/slugger/mstdnlambda/internal/payload"
)

func main() {
	flag.Parse()
	devenv.InitArgs()
	cfg.ParseConfig()
	logging.Reset()
	if !devenv.IsActive() {
		lambda.Start(handleRequest)
	} else {
		triggerDevEnv()
	}
}

func triggerDevEnv() {
	log := logging.GetLogForCategory(logging.DevEnvCategory)
	data, err := devenv.GetEventData()
	if err != nil {
		panic(fmt.Errorf("[GetEventData() failed] %w", err))
	}

	var event events.LambdaFunctionURLRequest
	err = json.Unmarshal(data, &event)
	if err != nil {
		panic(fmt.Errorf("[event unmarshal failed] %w", err))
	}

	resp, err := handleRequest(context.TODO(), event)
	if err != nil {
		panic(fmt.Errorf("[handleRequest failed] %w", err))
	} else {
		output, err := json.Marshal(resp)
		if err != nil {
			panic(fmt.Errorf("[resp marshal failed] %w", err))
		}
		log.Info(string(output))
	}
}

func handleRequest(ctx context.Context, event events.LambdaFunctionURLRequest) (*events.LambdaFunctionURLResponse, error) {
	log := logging.GetLogForCategory(logging.LambdaCategory)

	var vjwt *jwt.VerifiableJwt
	var req *payload.EncryptedPayload
	var err error

	logging.LogAsJson(log, logrus.DebugLevel, event, "event logged", false)

	if vjwt, err = http.ExtractJwt(event); err != nil {
		logging.LogAsJson(log, logrus.ErrorLevel, event, "jwt extract failed", true)
		e := fmt.Errorf("[jwt extract failed] %w", err)
		return nil, e
	}

	if req, err = http.ExtractPayload(event); err != nil {
		logging.LogAsJson(log, logrus.ErrorLevel, event, "payload extract failed", true)
		e := fmt.Errorf("[payload extract failed] %w", err)
		return nil, e
	}

	if !cfg.Cfg.IsSkipJwtVerify() {
		if err = jwt.Verify(vjwt); err != nil {
			logging.LogAsJson(log, logrus.ErrorLevel, event, "jwt verify failed", true)
			e := fmt.Errorf("[jwt verify failed] %w", err)
			return nil, e
		}
	} else {
		log.Warn("JWT VERIFICATION IS DISABLED!")
	}

	var msg string
	msg, err = payload.Decrypt(req)
	if err != nil {
		b64Payload := base64.StdEncoding.EncodeToString(req.Data)
		logging.LogAsJson(log.WithField("data", b64Payload), logrus.ErrorLevel, event, "payload decrypt failed", false)
		e := fmt.Errorf("[payload decrypt failed] %w", err)
		return nil, e
	}
	log.WithField("payload", msg).Debug("payload received")

	targets, err := http.ExtractTargets(event)
	if err != nil {
		logging.LogAsJson(log, logrus.ErrorLevel, event.RawPath, "targets extract failed", false)
		e := fmt.Errorf("[targets extract failed] %w", err)
		return nil, e
	}

	statusCode := 201
	statusTxt := "ok"
	for _, t := range targets {
		n := notify.New(t)
		if err = n.Send(msg); err != nil {
			e := fmt.Errorf("[notification failed] %w", err)
			logging.LogAsJson(log.WithField("err", e), logrus.ErrorLevel, n, "notification failed", false)
			statusCode = 500
			statusTxt = "fail"
			break
		}
	}

	return http.EncodeResponse(statusCode, statusTxt), nil
}

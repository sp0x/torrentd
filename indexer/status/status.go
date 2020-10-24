package status

import (
	"context"
	"fmt"
	"github.com/lileio/pubsub"
	"github.com/lileio/pubsub/middleware/defaults"
	"github.com/lileio/pubsub/providers/google"
	log "github.com/sirupsen/logrus"
)

const schemeTopic = "scrapescheme"
const schemeErrorTopic = "scheme.topic"
const LoginError = "login"
const TargetError = "bad-target"
const ContentError = "cant-fetch-content"
const Ok = "ok"

type ScrapeSchemeMessage struct {
	SchemeVersion string
	Site          string
	Code          string
	ResultsFound  int
}
type SchemeErrorMessage struct {
	Code          string
	Message       string
	Site          string
	SchemeVersion string
}

func SetupPubsub(projectId string) {
	provider, err := google.NewGoogleCloud(projectId)
	if err != nil {
		log.Errorf("%v\n", err)
		fmt.Printf("couldn't initialize google pubsub provider. status will not be published\n")
		return
	}
	//Service credentials exposed through: GOOGLE_APPLICATION_CREDENTIALS
	pubsub.SetClient(&pubsub.Client{
		ServiceName: "ihcph",
		Provider:    provider,
		Middleware:  defaults.Middleware,
	})
}

func PublishSchemeStatus(ctx context.Context, msg *ScrapeSchemeMessage) {
	log.
		WithFields(log.Fields{"message": msg}).
		Infof("Scheme %s: ", schemeTopic)
	pubsub.PublishJSON(ctx, schemeTopic, msg)
}

func PublishSchemeError(ctx context.Context, msg *SchemeErrorMessage) {
	log.
		WithFields(log.Fields{"message": msg}).
		Infof("ERR Scheme %s: ", schemeErrorTopic)
	pubsub.PublishJSON(ctx, schemeErrorTopic, msg)
}

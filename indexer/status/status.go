package status

import (
	"context"
	"fmt"
	"github.com/lileio/pubsub"
	"github.com/lileio/pubsub/middleware/defaults"
	"github.com/lileio/pubsub/providers/google"
	log "github.com/sirupsen/logrus"
	"os"
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
		log.Errorf("%v", err)
		fmt.Printf("couldn't initialize google pubsub provider")
		os.Exit(1)
	}
	//Service credentials exposed through: GOOGLE_APPLICATION_CREDENTIALS
	pubsub.SetClient(&pubsub.Client{
		ServiceName: "ihcph",
		Provider:    provider,
		Middleware:  defaults.Middleware,
	})
}

func PublishSchemeStatus(ctx context.Context, msg *ScrapeSchemeMessage) {
	pubsub.PublishJSON(ctx, schemeTopic, msg)
}

func PublishSchemeError(ctx context.Context, msg *SchemeErrorMessage) {
	pubsub.PublishJSON(ctx, schemeErrorTopic, msg)
}

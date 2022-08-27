package status

import (
	"context"

	"github.com/lileio/pubsub"
	"github.com/lileio/pubsub/middleware/defaults"
	"github.com/lileio/pubsub/providers/google"
	log "github.com/sirupsen/logrus"
)

const (
	schemeTopic      = "scrapescheme"
	schemeErrorTopic = "scheme.topic"
	LoginError       = "login"
	TargetError      = "bad-target"
	ContentError     = "cant-fetch-content"
	Ok               = "ok"
)

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

func SetupPubsub(projectID string) {
	provider, err := google.NewGoogleCloud(projectID)
	if err != nil {
		log.Errorf("%v\n", err)
		return
	}
	// Service credentials exposed through: GOOGLE_APPLICATION_CREDENTIALS
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

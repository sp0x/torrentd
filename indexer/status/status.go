package status

import (
	"context"
	"fmt"
	"github.com/lileio/pubsub"
	"github.com/lileio/pubsub/middleware/defaults"
	"github.com/lileio/pubsub/providers/google"
	"github.com/sp0x/torrentd/indexer"
	"github.com/spf13/viper"
	"os"
)

const schemeTopic = "scrapecheme"
const schemeErrorTopic = "scheme.topic"
const LoginError = "login"
const TargetError = "bad-target"
const ContentError = "cant-fetch-content"
const Ok = "ok"

type scrapeSchemeMessage struct {
	SchemeVersion string
	Site          string
	Code          string
}
type schemeErrorMessage struct {
	Code          string
	Message       string
	Site          string
	SchemeVersion string
}

func setupPubsubConfig() {
	viper.AutomaticEnv()
	_ = viper.BindEnv("firebase_project")
	_ = viper.BindEnv("firebase_credentials_file")
}

func init() {
	setupPubsubConfig()
	projectId := viper.GetString("firebase_project")
	provider, err := google.NewGoogleCloud(projectId)
	if err != nil {
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

func PublishSchemeStatus(ctx context.Context, statusCode string, definition *indexer.IndexerDefinition) {
	msg := scrapeSchemeMessage{
		Code:          statusCode,
		Site:          definition.Site,
		SchemeVersion: definition.Version,
	}
	pubsub.PublishJSON(ctx, schemeTopic, msg)
}

func PublishSchemeError(ctx context.Context, errorCode string, err error, definition *indexer.IndexerDefinition) {
	msg := &schemeErrorMessage{
		Code:          errorCode,
		Site:          definition.Site,
		SchemeVersion: definition.Version,
		Message:       fmt.Sprintf("couldn't log in: %s", err),
	}
	pubsub.PublishJSON(ctx, schemeErrorTopic, msg)
}

package indexer

import (
	"github.com/sirupsen/logrus"
)

//Read anything from the content that's needed
//so we can extract info about our run
func setupContext(r *Runner, ctx *RunContext, dom RawScrapeItem) {
	//ctx.SearchKeywords.DOM = dom
	for _, item := range r.definition.Search.Context {
		r.logger.
			WithFields(logrus.Fields{"block": item.Block.String()}).
			Debugf("Extracting context field %q", item.Field)

		val, err := item.Block.Match(dom)
		if err != nil {
			r.logger.
				WithFields(logrus.Fields{"block": item.Block.String()}).
				Debugf("Failed while extracting context field %q", item.Field)
			continue
		}
		if item.Field == "searchId" {
			ctx.Search.SetId(val.(string))
		}

	}
}

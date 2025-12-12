package okta

import (
	"net/url"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/sirupsen/logrus"
)

func (p *oktaProvider) GetNextTokenFromResponse(resp *okta.Response) string {

	nextPageURL, err := url.Parse(resp.NextPage)
	if err != nil {
		logrus.Warnf("Failed to parse next page URL: %v", err)
	} else {
		q := nextPageURL.Query()
		if after := q.Get("after"); len(after) > 0 {
			return after
		}
	}

	return ""

}

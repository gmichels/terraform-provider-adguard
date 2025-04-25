package adguard

import (
	"github.com/gmichels/adguard-client-go"
	adgmodels "github.com/gmichels/adguard-client-go/models"
)

// GetRewrite - Return a DNS rewrite rule based on the domain and answer
func GetRewrite(adg *adguard.ADG, domain, answer string) (*adgmodels.RewriteEntry, error) {
	// retrieve all DNS rewrite rules
	allRewrites, err := adg.RewriteList()
	if err != nil {
		return nil, err
	}

	// loop over the results until we find the one we want
	for _, rewrite := range *allRewrites {
		if rewrite.Domain == domain && rewrite.Answer == answer {
			return &rewrite, nil
		}
	}

	// when no matches are found
	return nil, nil
}

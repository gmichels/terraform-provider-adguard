package adguard

import (
	"github.com/gmichels/adguard-client-go"
	adgmodels "github.com/gmichels/adguard-client-go/models"
)

// GetListFilterById - Returns a list filter based on its name and whether it's a whitelist filter
func GetListFilterById(adg *adguard.ADG, id int64) (*adgmodels.Filter, bool, error) {
	allFilters, err := adg.FilteringStatus()
	if err != nil {
		return nil, false, err
	}

	// go through the blacklist filters in the response until we find the one we want
	for _, filterInfo := range allFilters.Filters {
		if filterInfo.Id == id {
			return &filterInfo, false, nil
		}
	}
	// go through the whitelist filters in the response until we find the one we want
	for _, filterInfo := range allFilters.WhitelistFilters {
		if filterInfo.Id == id {
			return &filterInfo, true, nil
		}
	}

	// when no matches are found
	return nil, false, nil
}

// GetListFilterByName - Returns a list filter based on its name and whether it's a whitelist filter
func GetListFilterByName(adg *adguard.ADG, listName string) (*adgmodels.Filter, bool, error) {
	allFilters, err := adg.FilteringStatus()
	if err != nil {
		return nil, false, err
	}

	// go through the blacklist filters in the response until we find the one we want
	for _, filterInfo := range allFilters.Filters {
		if filterInfo.Name == listName {
			return &filterInfo, false, nil
		}
	}
	// go through the whitelist filters in the response until we find the one we want
	for _, filterInfo := range allFilters.WhitelistFilters {
		if filterInfo.Name == listName {
			return &filterInfo, true, nil
		}
	}

	// when no matches are found
	return nil, false, nil
}

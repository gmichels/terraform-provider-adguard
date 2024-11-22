package adguard

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// define a cache for static results from ADG
var apiCache = struct {
	mu     sync.Mutex
	values map[string][]string
}{
	values: make(map[string][]string),
}

// retrieve an entry from the cache based on a key
func getFromCache(key string) []string {
	// ensure we are thread-safe
	apiCache.mu.Lock()
	defer apiCache.mu.Unlock()

	// return cached values, if there
	if cached, exists := apiCache.values[key]; exists {
		return cached
	}

	return nil
}

// check if a slice contains a string
func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

// converts an array of string to array of attr.Value of StringType
func convertToAttr(elems []string) []attr.Value {
	var output []attr.Value

	for _, item := range elems {
		output = append(output, types.StringValue(item))
	}
	return output
}

// given a duration in milliseconds, convert to a HH:MM string
func convertMsToHourMinutes(duration_ms int64) string {
	// convert provided interval in ms to duration
	d := time.Duration(duration_ms * 1000000)
	// extract hours/minutes
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	return fmt.Sprintf("%02d:%02d", h, m)
}

// given a HH:MM string, convert to a duration in milliseconds
func convertHoursMinutesToMs(duration_hhmm string) int64 {
	d, _ := time.ParseDuration(strings.Replace(duration_hhmm, ":", "h", -1) + "m")

	return d.Milliseconds()
}

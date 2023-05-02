package adguard

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

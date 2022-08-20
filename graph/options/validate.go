package options

import (
	"github.com/zalgonoise/x/graph/errs"
)

// adjancy lists will not hold edge weights
func (c *GraphConfig) validateAdjancyListConfig() bool {
	if c.GraphType != GraphList {
		return true
	}
	if !c.IsUnweighted {
		return false
	}
	return true
}

func (c *GraphConfig) Validate() (bool, error) {
	if !c.validateAdjancyListConfig() {
		return false, errs.InvalidAdjList
	}
	// more validations if needed

	return true, nil
}

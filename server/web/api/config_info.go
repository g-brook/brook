package api

import (
	"github.com/brook/server/web/db"
	"github.com/brook/server/web/errs"
)

func init() {
	RegisterRoute(NewRouteNotAuth("/getConfigs", "POST"), getConfigs)
	RegisterRoute(NewRouteNotAuth("/addConfigs", "POST"), addConfigs)
}

const ConfigKey = "brook_config"

// getConfigs retrieves configuration information from the database
// It takes a pointer to a Request with any type as parameter
// and returns a Response containing the configuration data or error information
func getConfigs(*Request[any]) *Response {
	infos, err := db.Get[[]ConfigInfo](ConfigKey)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "Get Config is err")
	}
	return NewResponseSuccess(infos)
}

func addConfigs(*Request[ConfigInfo]) *Response {
	return NewResponseSuccess(nil)
}

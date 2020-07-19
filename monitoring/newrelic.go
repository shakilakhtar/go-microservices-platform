package monitoring

import (
	logger "github.com/sirupsen/logrus"
	kiterrors "shakilakhtar/go-microservices-platform/errors"
	"fmt"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/newrelic/go-agent"
	"os"
)

const lincenseKey = "licenseKey"

//SetupNewRelic - Utility for creating a new application with new relic
//This is spawn a new go routine
func SetupNewRelic(appName string, licenseKey string) (newrelic.Application, error) {

	if licenseKey != "" {
		config := newrelic.NewConfig(appName, licenseKey)
		config.Enabled = false
		app, err := newrelic.NewApplication(config)
		if err != nil {
			logger.Crit("Could not start NewRelic application", "error", err)
			return err
		}
		return app, err

	}
	return

}

//SetupNewRelicFromCFEnv - Utility for Enabling/Disabling monitoring tool by reading from environment variables
func SetupNewRelicFromCFEnv() error {
	newRelicServiceInstanceName := os.Getenv("NEW_RELIC_INSTANCE_NAME")

	if newRelicServiceInstanceName == "" {
		logger.Error("New Relic Service instance name not set")
		return kiterrors.NewError("newRelicServiceInstanceName not set")
	}

	appEnv, err := cfenv.Current()
	if err != nil {
		logger.Crit("CF environment variable is not set, shutting down", "error", err)
		return err
	}

	if newRelicServiceInstanceName != "" {
		service, err := appEnv.Services.WithName(newRelicServiceInstanceName)
		if err != nil {
			logger.Error("error getting newrelic service instance: ", "Error", err.Error())
			return err
		}
		if service != nil {
			appNameNode := fmt.Sprintf("%s (%d)", appEnv.ApplicationURIs[0], appEnv.Index)
			SetupNewRelic(appNameNode, service.Credentials[lincenseKey].(string))
		}
	} else {
		logger.Error("New Relic Service instance name not set ")
	}
	return nil
}

package cf

import (
	gokiterrors "shakilakhtar/go-microservices-platform/errors"
	"github.com/cloudfoundry-community/go-cfenv"
)

//GetVCAPServiceProperty retrieves a property from bound service credentials.
func GetVCAPServiceProperty(serviceName string, propertyName string, appEnv *cfenv.App) (propertyValue string, err error) {
	if propertyName == "" {
		return "", gokiterrors.NewError("Must supply a property name value.")
	}

	service, err := getVCAPService(serviceName, appEnv)
	if err != nil {
		return "", err
	}

	propertyValue = service.Credentials[propertyName].(string)
	if propertyValue == "" {
		return "", gokiterrors.NewError("Error retrieving property: " + propertyName)
	}
	return
}

func getVCAPService(serviceName string, appEnv *cfenv.App) (service *cfenv.Service, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = gokiterrors.NewError("Error retrieving VCAP Service: ")
		}
	}()

	service, err = appEnv.Services.WithName(serviceName)
	return
}

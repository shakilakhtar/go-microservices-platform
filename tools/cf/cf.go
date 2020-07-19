package cf

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// DefaultTimeout is a configurable property for when the command line session ends.
var DefaultTimeout = 180 * time.Second

// ShouldExitSuccessfully checks that the session exited gracefully.
func ShouldExitSuccessfully(session *gexec.Session) {
	Eventually(session, DefaultTimeout).Should(gexec.Exit())
	Expect(session.ExitCode()).To(Equal(0))
}

// ShellOut starts a command line session.
func ShellOut(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func cf(args ...string) *gexec.Session {
	return helpers.Run("cf", args...)
}

// PushAppFromPath uses PushAppFromPathWithFlags method, passing an empty slice
// of flags as to bypass the requirement for flags.
func PushAppFromPath(path, appName string) {
	PushAppFromPathWithFlags(path, appName, []string{}...)
}

// PushAppFromPathWithFlags uses the Cloud Foundry CLI and the running command
// line session to push an application with the given flags.
func PushAppFromPathWithFlags(path, appName string, flags ...string) {
	currentDir, err := os.Getwd()
	err = os.Chdir(path)
	Expect(err).ShouldNot(HaveOccurred())

	pushAppWithArgs := append([]string{"push", appName}, flags...)
	pushSession := cf(pushAppWithArgs...)
	ShouldExitSuccessfully(pushSession)

	err = os.Chdir(currentDir)
	Expect(err).ShouldNot(HaveOccurred())
}

// GetCredentialsFromAppEnv returns json representation of credentials from
// app's vcaps. Nested objects within credentials are not supported.
func GetCredentialsFromAppEnv(appName, serviceBrokerName string) (string, error) {
	envString := GetAppEnvironment(appName)
	patternWithServiceBroker := "s*\"" + serviceBrokerName + "\":\\s\\[\\s*{\\s*\\n*\"credentials\":\\s*({\\n\\s*\"ingest\"[^}]+},\\n\\s*\"query\"[^}]+}\\n\\s*})"
	re := regexp.MustCompile(patternWithServiceBroker)
	match := re.FindStringSubmatch(envString)
	if match != nil && len(match) != 0 {
		patternWithOnlyCredentials := "s*\\n*({\\n\\s*\"ingest\"[^}]+},\\n\\s*\"query\"[^}]+}\\n\\s*})"
		re = regexp.MustCompile(patternWithOnlyCredentials)
		credentialsMatch := re.FindStringSubmatch(match[1])
		if credentialsMatch != nil && len(credentialsMatch) != 0 {
			return credentialsMatch[1], nil
		}
	}
	return "", errors.New("failed to get credentials from app environment for the pattern:" + patternWithServiceBroker)
}

// RemoveBroker uses the Cloud Foundry CLI and the running command
// line session to remove a service broker.
func RemoveBroker(brokerName, brokerAppName string) {
	purgeServiceOfferingSession := cf("purge-service-offering", brokerName, "-f")
	ShouldExitSuccessfully(purgeServiceOfferingSession)

	deleteBrokerSession := cf("delete-service-broker", brokerName, "-f")
	ShouldExitSuccessfully(deleteBrokerSession)

	DeleteApp(brokerAppName)
}

// DeleteApp uses the Cloud Foundry CLI and the running command
// line session to delete an application.
func DeleteApp(appName string) {
	deleteSession := cf("delete", appName, "-f", "-r")
	ShouldExitSuccessfully(deleteSession)
}

// UpdateCups uses the Cloud Foundry CLI and the running command
// line session to update a CF custom user provided service.
func UpdateCups(servicename, credential string) {
	cupsSession := cf("uups",
		servicename,
		"-p", credential,
	)
	ShouldExitSuccessfully(cupsSession)
}

// SetupCups uses the Cloud Foundry CLI and the running command
// line session to create a CF custom user provided service.
func SetupCups(servicename, credential string) {
	cupsSession := cf("cups",
		servicename,
		"-p", credential,
	)
	ShouldExitSuccessfully(cupsSession)
}

// CreateServiceInstance uses the Cloud Foundry CLI and the running command
// line session to create a service instance from the marketplace.
func CreateServiceInstance(serviceName, servicePlan, serviceInstanceName string) {
	createServiceInstanceSession := cf("create-service", serviceName, servicePlan, serviceInstanceName)
	ShouldExitSuccessfully(createServiceInstanceSession)
}

// CreateServiceInstanceWithConfig uses the Cloud Foundry CLI and the running command
// line session to create a service instance from the marketplace with given config.
func CreateServiceInstanceWithConfig(serviceName, servicePlan, serviceInstanceName, config string) {
	createServiceInstanceSession := cf("create-service", serviceName, servicePlan, serviceInstanceName, "-c", config)
	ShouldExitSuccessfully(createServiceInstanceSession)
}

// GetServiceInstance uses the Cloud Foundry CLI and the running command line
// session to get a service instance in the current CF space.
func GetServiceInstance(serviceInstanceName string) {
	displayServiceInstanceSession := cf("service", serviceInstanceName)
	ShouldExitSuccessfully(displayServiceInstanceSession)
}

// IsServiceDeployed uses the Cloud Foundry CLI and the running command line
// session to check if a given service is deployed in the current CF space.
func IsServiceDeployed(serviceName string) bool {
	session := cf("service", serviceName)
	Eventually(session, DefaultTimeout).Should(gexec.Exit())
	if session.ExitCode() == 0 {
		return true
	}
	return false
}

// DeleteServiceInstance uses the Cloud Foundry CLI and the running command
// line session to delete a service instance from the current CF space.
func DeleteServiceInstance(serviceName string) {
	deleteServiceInstance := cf("delete-service", serviceName, "-f")
	ShouldExitSuccessfully(deleteServiceInstance)
}

// UnbindService uses the Cloud Foundry CLI and the running command
// line session to unbind a service instance from an application in the current
// CF space.
func UnbindService(serviceName, appName string) {
	unbindServiceInstanceSession := cf("unbind-service", appName, serviceName)
	ShouldExitSuccessfully(unbindServiceInstanceSession)
}

// BindService uses the Cloud Foundry CLI and the running command
// line session to bind a service instance from an application in the current
// CF space.
func BindService(appName string, serviceNames ...string) {
	for _, serviceName := range serviceNames {
		bindServiceInstanceSession := cf("bind-service", appName, serviceName)
		ShouldExitSuccessfully(bindServiceInstanceSession)
	}

	restageAppSession := cf("restart", appName)
	ShouldExitSuccessfully(restageAppSession)
}

// GetAppEnvironment uses the Cloud Foundry CLI and the running command line
// session to get the application environment variables in the currrent CF space.
func GetAppEnvironment(appName string) string {
	envDetails := cf("env", appName)
	ShouldExitSuccessfully(envDetails)
	return string(envDetails.Out.Contents())
}

// GetAppLogs uses the Cloud Foundry CLI and the running command line
// session to get the recent application logs in the currrent CF space.
func GetAppLogs(appName string) string {
	logDetails := cf("logs", appName, "--recent")
	ShouldExitSuccessfully(logDetails)
	return string(logDetails.Out.Contents())
}

// GetServiceGUID uses the Cloud Foundry CLI and the running command line
// session to get the service instance guid.
func GetServiceGUID(serviceName string) string {
	guid := cf("service", serviceName, "--guid")
	ShouldExitSuccessfully(guid)
	return string(guid.Out.Contents())
}

// GetAppStatus uses the Cloud Foundry CLI and the running command line
// session to get the application status in the currrent CF space.
func GetAppStatus(appName string) string {
	statusDetails := cf("app", appName)
	ShouldExitSuccessfully(statusDetails)
	return string(statusDetails.Out.Contents())
}

// DisplayMarketplaceOffering uses the Cloud Foundry CLI and the running command
// line session to list all available services in the marketplace.
func DisplayMarketplaceOffering(serviceName string) {
	marketplaceSession := cf("m", "-s", serviceName)
	ShouldExitSuccessfully(marketplaceSession)
}

// RestageApp uses the Cloud Foundry CLI and the running command line session to
// restage an application in the current CF space.
func RestageApp(appName string) {
	restageAppSession := cf("restage", appName)
	ShouldExitSuccessfully(restageAppSession)
}

// CreateServiceBroker uses the Cloud Foundry CLI and the running command
// line session to create a service broker in the current CF space.
func CreateServiceBroker(serviceName, userName, password, url string) {
	createBrokerSession := cf("create-service-broker", serviceName, userName, password, url)
	ShouldExitSuccessfully(createBrokerSession)
	enableBrokerSession := cf("enable-service-access", serviceName)
	ShouldExitSuccessfully(enableBrokerSession)
}

// Login uses the Cloud Foundry CLI and the running command line session to
// login with given credentials.
func Login(host, userName, password string) {
	loginSession := cf("login", "-a", "https://api."+host, "-u", userName, "-p", password, "--skip-ssl-validation")
	ShouldExitSuccessfully(loginSession)
}

// LoginWithOrgAndSpace uses the Cloud Foundry CLI and the running command
// line session to login with given credentials to the desired CF space and org.
func LoginWithOrgAndSpace(host, userName, password, org, space string) {
	loginSession := cf("login", "-a", "https://api."+host, "-u", userName, "-p", password, "-o", org, "-s", space, "--skip-ssl-validation")
	ShouldExitSuccessfully(loginSession)
}

func checkRoute(host, domain string) {
	checkRouteSession := cf("check-route", host, domain)
	ShouldExitSuccessfully(checkRouteSession)
	Expect(string(checkRouteSession.Out.Contents())).ShouldNot(ContainSubstring("does not exist"))
}

//SetEnv sets an environment variable to an app on Cloud Foundry
func SetEnv(appName, envKey, envValue string) {
	session := cf("set-env", appName, envKey, envValue)
	ShouldExitSuccessfully(session)
}

//RestartApp restarts the app on Cloud Foundry
func RestartApp(appName string) {
	session := cf("restart", appName)
	ShouldExitSuccessfully(session)
}

//IsAppDeployed checks if the given appName is deployed on Cloud Foundry
func IsAppDeployed(appName string) bool {
	session := cf("app", appName)
	Eventually(session, DefaultTimeout).Should(gexec.Exit())
	if session.ExitCode() == 0 {
		return true
	}
	return false
}

//BindServiceWithoutRestage binds the appName to a service instances on Cloud Foundry
func BindServiceWithoutRestage(appName string, serviceNames ...string) {
	for _, serviceName := range serviceNames {
		bindServiceInstanceSession := cf("bind-service", appName, serviceName)
		ShouldExitSuccessfully(bindServiceInstanceSession)
	}
}

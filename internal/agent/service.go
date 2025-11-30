package agent

import (
	"os"

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	config "github.com/thand-io/agent/internal/config"
)

// ServiceProgram implements the service.Interface
type ServiceProgram struct {
	exit   chan struct{}
	config *config.Config
}

func (p *ServiceProgram) Start(s service.Service) error {
	logrus.Infoln("Thand Agent service starting")
	go p.run()
	return nil
}

func (p *ServiceProgram) run() {
	// Start the agent web service
	_, err := StartWebService(p.config)

	if err != nil {
		logrus.WithError(err).Errorf("Failed to start web service")
		return
	}

	logrus.Infoln("Thand Agent service is running")
}

func (p *ServiceProgram) Stop(s service.Service) error {
	logrus.Infoln("Thand Agent service stopping")
	close(p.exit)
	return nil
}

// createService creates a new service instance
func CreateService(cfg *config.Config) (service.Service, error) {
	svcConfig := getServiceConfig()

	prg := &ServiceProgram{
		exit:   make(chan struct{}),
		config: cfg,
	}

	return service.New(prg, svcConfig)
}

// getServiceConfig returns the service configuration
func getServiceConfig() *service.Config {
	exePath, err := os.Executable()

	if err != nil {
		logrus.Fatal(err)
	}

	return &service.Config{
		Name:        "thand",
		DisplayName: "Thand Agent Service",
		Description: "Thand Agent - Just-in-time access to cloud infrastructure and SaaS applications",
		Executable:  exePath,
		Arguments: []string{
			"agent", // Runs the web server
		},
	}
}

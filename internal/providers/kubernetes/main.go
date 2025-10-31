package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/sirupsen/logrus"

	"github.com/thand-io/agent/internal/models"
	"github.com/thand-io/agent/internal/providers"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const ProviderName = "kubernetes"

// kubernetesProvider implements the ProviderImpl interface for Kubernetes
type kubernetesProvider struct {
	*models.BaseProvider
	client           kubernetes.Interface
	permissions      []models.ProviderPermission
	permissionsIndex bleve.Index
	roles            []models.ProviderRole
	rolesIndex       bleve.Index
}

func (p *kubernetesProvider) Initialize(provider models.Provider) error {
	p.BaseProvider = models.NewBaseProvider(
		provider,
		models.ProviderCapabilityRBAC,
	)

	// Initialize Kubernetes client
	config, err := p.getKubernetesConfig()
	if err != nil {
		return fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	p.client = client

	// Load Kubernetes permissions and roles
	err = p.LoadPermissions()
	if err != nil {
		return fmt.Errorf("failed to load permissions: %w", err)
	}

	err = p.LoadRoles()
	if err != nil {
		return fmt.Errorf("failed to load roles: %w", err)
	}

	return nil
}

func (p *kubernetesProvider) GetClient() kubernetes.Interface {
	return p.client
}

// getKubernetesConfig returns the appropriate Kubernetes configuration
func (p *kubernetesProvider) getKubernetesConfig() (*rest.Config, error) {
	// Try in-cluster config first (when running inside K8s)
	if config, err := rest.InClusterConfig(); err == nil {
		logrus.Info("Using in-cluster Kubernetes configuration")
		return config, nil
	}

	// Fallback to kubeconfig file
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	logrus.WithField("kubeconfig", kubeconfigPath).Info("Using kubeconfig file")
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

func init() {
	providers.Register(ProviderName, &kubernetesProvider{})
}

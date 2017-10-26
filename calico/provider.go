package calico

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/projectcalico/libcalico-go/lib/api"
	"github.com/projectcalico/libcalico-go/lib/client"
)

// Provider is the provider for terraform
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"datastore_type": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     schema.EnvDefaultFunc("CALICO_DATASTORE_TYPE", "etcdv2"),
				Description: "Indicates the datastore to use (required for Kubernetes as the default is etcdv2)",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					datastoreType := v.(string)
					if datastoreType != "etcdv2" && datastoreType != "kubernetes" {
						errors = append(errors, fmt.Errorf("%q: %s", k, "etcdv2 and kubernetes are the only supported values"))
					}
					return
				},
			},
			"etcd_endpoints": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_BACKEND_ETCD_ENDPOINTS", ""),
				Description: "multiple etcd endpoints separated by comma",
			},
			"etcd_username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_BACKEND_ETCD_USERNAME", ""),
				Description: "Etcd username",
			},
			"etcd_password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_BACKEND_ETCD_PASSWORD", ""),
				Description: "Etcd password",
			},
			"etcd_key_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_BACKEND_ETCD_ETCD_KEY_FILE", ""),
				Description: "File location keyfile",
			},
			"etcd_cert_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_BACKEND_ETCD_ETCD_CERT_FILE", ""),
				Description: "File location certfile",
			},
			"etcd_ca_cert_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_BACKEND_ETCD_ETCD_CA_CERT_FILE", ""),
				Description: "File location cacert",
			},
			"kubeconfig": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_KUBECONFIG", ""),
				Description: "K8sKubeconfigFile`",
			},
			"k8s_api_endpoint": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_K8S_API_ENDPOINT", ""),
				Description: "K8sServer",
			},
			"k8s_cert_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_K8S_CERT_FILE", ""),
				Description: "K8sClientCertificate",
			},
			"k8s_key_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_K8S_KEY_FILE", ""),
				Description: "K8sClientKey",
			},
			"k8s_ca_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_K8S_CA_FILE", ""),
				Description: "K8sCertificateAuthority",
			},
			"k8s_token": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CALICO_K8S_TOKEN", ""),
				Description: "K8sToken",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"calico_hostendpoint": resourceCalicoHostendpoint(),
			"calico_profile":      resourceCalicoProfile(),
			"calico_policy":       resourceCalicoPolicy(),
			"calico_ippool":       resourceCalicoIpPool(),
			"calico_bgppeer":      resourceCalicoBgpPeer(),
			"calico_node":         resourceCalicoNode(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	calicoConfig := api.CalicoAPIConfig{}

	datastoreType := d.Get("datastore_type").(string)

	switch datastoreType {
	case "etcdv2":
		calicoConfig.Spec.DatastoreType = api.DatastoreType(datastoreType)

		calicoConfig.Spec.EtcdEndpoints = d.Get("etcd_endpoints").(string)
		calicoConfig.Spec.EtcdUsername = d.Get("etcd_username").(string)
		calicoConfig.Spec.EtcdPassword = d.Get("etcd_password").(string)
		calicoConfig.Spec.EtcdKeyFile = d.Get("etcd_key_file").(string)
		calicoConfig.Spec.EtcdCertFile = d.Get("etcd_cert_file").(string)
		calicoConfig.Spec.EtcdCACertFile = d.Get("etcd_ca_cert_file").(string)
	case "kubernetes":
		calicoConfig.Spec.DatastoreType = api.DatastoreType(datastoreType)

		calicoConfig.Spec.Kubeconfig = d.Get("kubeconfig").(string)
		calicoConfig.Spec.K8sAPIEndpoint = d.Get("k8s_api_endpoint").(string)
		calicoConfig.Spec.K8sCertFile = d.Get("k8s_cert_file").(string)
		calicoConfig.Spec.K8sKeyFile = d.Get("k8s_key_file").(string)
		calicoConfig.Spec.K8sCAFile = d.Get("k8s_ca_file").(string)
		calicoConfig.Spec.K8sAPIToken = d.Get("k8s_token").(string)
	}

	calicoClient, err := client.New(calicoConfig)
	if err != nil {
		return nil, err
	}

	config := config{
		config: calicoConfig,
		Client: calicoClient,
	}

	log.Printf("Configured: %#v", config)

	if err := config.loadAndValidate(); err != nil {
		return nil, err
	}

	return config, nil
}

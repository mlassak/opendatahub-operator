/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	kfdeftypes "github.com/opendatahub-io/opendatahub-operator/apis/kfdef.apps.kubeflow.org/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:openapi-gen=true
// Placeholder for the plugin API.
type KfGcpPlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GcpPluginSpec `json:"spec,omitempty"`
}

// GcpPluginSpec defines the desired state of GcpPlugin
type GcpPluginSpec struct {
	Auth *Auth `json:"auth,omitempty"`

	// SAClientId if supplied grant this service account cluster admin access
	// TODO(jlewi): Might want to make it a list
	SAClientId string `json:"username,omitempty"`

	// CreatePipelinePersistentStorage indicates whether to create storage.
	// Use a pointer so we can distinguish unset values.
	CreatePipelinePersistentStorage *bool `json:"createPipelinePersistentStorage,omitempty"`

	// EnableWorkloadIdentity indicates whether to enable workload identity.
	// Use a pointer so we can distinguish unset values.
	EnableWorkloadIdentity *bool `json:"enableWorkloadIdentity,omitempty"`

	// DeploymentManagerConfig provides location of the deployment manager configs.
	DeploymentManagerConfig *DeploymentManagerConfig `json:"deploymentManagerConfig,omitempty"`

	Project         string `json:"project,omitempty"`
	Email           string `json:"email,omitempty"`
	IpName          string `json:"ipName,omitempty"`
	Hostname        string `json:"hostname,omitempty"`
	Zone            string `json:"zone,omitempty"`
	UseBasicAuth    bool   `json:"useBasicAuth"`
	SkipInitProject bool   `json:"skipInitProject,omitempty"`
	DeleteStorage   bool   `json:"deleteStorage,omitempty"`
}
type Auth struct {
	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	IAP       *IAP       `json:"iap,omitempty"`
}

type BasicAuth struct {
	Username string                `json:"username,omitempty"`
	Password *kfdeftypes.SecretRef `json:"password,omitempty"`
}

type IAP struct {
	OAuthClientId     string                `json:"oAuthClientId,omitempty"`
	OAuthClientSecret *kfdeftypes.SecretRef `json:"oAuthClientSecret,omitempty"`
}

type DeploymentManagerConfig struct {
	RepoRef *kfdeftypes.RepoRef `json:"repoRef,omitempty"`
}

// IsValid returns true if the spec is a valid and complete spec.
// If false it will also return a string providing a message about why its invalid.
func (s *GcpPluginSpec) IsValid() (bool, string) {
	if len(s.Hostname) > 63 {
		return false, fmt.Sprintf("Invaid host name: host name %s is longer than 63 characters. Please shorten the metadata.name.", s.Hostname)
	}
	basicAuthSet := s.Auth.BasicAuth != nil
	iapAuthSet := s.Auth.IAP != nil

	if basicAuthSet == iapAuthSet {
		return false, "Exactly one of BasicAuth and IAP must be set; the other should be nil"
	}

	if basicAuthSet {
		msg := ""

		isValid := true

		if s.Auth.BasicAuth.Username == "" {
			isValid = false
			msg += "BasicAuth requires username. "
		}

		if s.Auth.BasicAuth.Password == nil {
			isValid = false
			msg += "BasicAuth requires password. "
		}

		return isValid, msg
	}

	if iapAuthSet {
		msg := ""
		isValid := true

		if s.Auth.IAP.OAuthClientId == "" {
			isValid = false
			msg += "IAP requires OAuthClientId. "
		}

		if s.Auth.IAP.OAuthClientSecret == nil {
			isValid = false
			msg += "IAP requires OAuthClientSecret. "
		}

		return isValid, msg
	}

	return false, "Either BasicAuth or IAP must be set"
}

func (p *GcpPluginSpec) GetCreatePipelinePersistentStorage() bool {
	if p.CreatePipelinePersistentStorage == nil {
		return true
	}

	v := p.CreatePipelinePersistentStorage
	return *v
}

func (p *GcpPluginSpec) GetEnableWorkloadIdentity() bool {
	if p.EnableWorkloadIdentity == nil {
		return true
	}

	v := p.EnableWorkloadIdentity
	return *v
}

// GcpPluginStatus defines the observed state of GcpPlugin
type GcpPluginStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GcpPlugin is the Schema for the gcpplugins API
type GcpPlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpPluginSpec   `json:"spec,omitempty"`
	Status GcpPluginStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GcpPluginList contains a list of GcpPlugin
type GcpPluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpPlugin `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpPlugin{}, &GcpPluginList{})
}

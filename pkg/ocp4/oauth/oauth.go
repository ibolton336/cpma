package oauth

import (
	"github.com/fusor/cpma/internal/io"
	"github.com/fusor/cpma/pkg/ocp4/secrets"
	configv1 "github.com/openshift/api/legacyconfig/v1"
	oauthv1 "github.com/openshift/api/oauth/v1"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	// TODO: Is this line needed at all? It may be superflous to
	// ocp3.go/init()/configv1.InstallLegacy(scheme.Scheme)
	oauthv1.Install(scheme.Scheme)
}

// reference:
//   [v3] OCPv3:
//   - [1] https://docs.openshift.com/container-platform/3.11/install_config/configuring_authentication.html#identity_providers_master_config
//   [v4] OCPv4:
//   - [2] htpasswd: https://docs.openshift.com/container-platform/4.0/authentication/understanding-identity-provider.html
//   - [3] github: https://docs.openshift.com/container-platform/4.0/authentication/identity_providers/configuring-github-identity-provider.html

// Shared CRD part, present in all types of OAuth CRDs
type OAuthCRD struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   MetaData `yaml:"metadata"`
	Spec       struct {
		IdentityProviders []interface{} `yaml:"identityProviders"`
	} `yaml:"spec"`
}

type identityProviderCommon struct {
	Name          string `yaml:"name"`
	Challenge     bool   `yaml:"challenge"`
	Login         bool   `yaml:"login"`
	MappingMethod string `yaml:"mappingMethod"`
	Type          string `yaml:"type"`
}

type MetaData struct {
	Name      string `yaml:"name"`
	NameSpace string `yaml:"namespace"`
}

var (
	APIVersion = "config.openshift.io/v1"
	// GetFile allows to mock file retrieval
	GetFile = io.GetFile
)

// Transform converts OCPv3 OAuth to OCPv4 OAuth Custom Resources
func Transform(oauthconfig *configv1.OAuthConfig) (*OAuthCRD, []secrets.Secret, error) {
	var auth = oauthconfig.DeepCopy()
	var err error

	var oauthCrd OAuthCRD
	oauthCrd.APIVersion = APIVersion
	oauthCrd.Kind = "OAuth"
	oauthCrd.Metadata.Name = "cluster"
	oauthCrd.Metadata.NameSpace = "openshift-config"

	var idP interface{}
	var secretsSlice []secrets.Secret

	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	for _, p := range auth.IdentityProviders {
		secret := secrets.Secret{}
		certSecret := secrets.Secret{}
		keySecret := secrets.Secret{}

		p.Provider.Object, _, err = serializer.Decode(p.Provider.Raw, nil, nil)
		if err != nil {
			return nil, nil, err
		}

		switch kind := p.Provider.Object.GetObjectKind().GroupVersionKind().Kind; kind {
		case "GitHubIdentityProvider":
			idP, secret = buildGitHubIP(serializer, p)
		case "GitLabIdentityProvider":
			idP, secret = buildGitLabIP(serializer, p)
		case "GoogleIdentityProvider":
			idP, secret = buildGoogleIP(serializer, p)
		case "HTPasswdPasswordIdentityProvider":
			idP, secret = buildHTPasswdIP(serializer, p)
		case "OpenIDIdentityProvider":
			idP, secret = buildOpenIDIP(serializer, p)
		case "RequestHeaderIdentityProvider":
			idP = buildRequestHeaderIP(serializer, p)
		case "LDAPPasswordIdentityProvider":
			idP = buildLdapIP(serializer, p)
		case "KeystonePasswordIdentityProvider":
			idP, certSecret, keySecret = buildKeystoneIP(serializer, p)
			if certSecret != (secrets.Secret{}) {
				secretsSlice = append(secretsSlice, certSecret)
				secretsSlice = append(secretsSlice, keySecret)
			}
		case "BasicAuthPasswordIdentityProvider":
			idP, certSecret, keySecret = buildBasicAuthIP(serializer, p)
			if certSecret != (secrets.Secret{}) {
				secretsSlice = append(secretsSlice, certSecret)
				secretsSlice = append(secretsSlice, keySecret)
			}
		default:
			logrus.Infof("Can't handle %s OAuth kind", kind)
		}
		oauthCrd.Spec.IdentityProviders = append(oauthCrd.Spec.IdentityProviders, idP)
		if secret != (secrets.Secret{}) {
			secretsSlice = append(secretsSlice, secret)
		}
	}

	return &oauthCrd, secretsSlice, nil
}

// GenYAML returns a YAML of the OAuthCRD
func (oauth *OAuthCRD) GenYAML() []byte {
	yamlBytes, err := yaml.Marshal(&oauth)
	if err != nil {
		logrus.Fatal(err)
	}

	return yamlBytes
}

package ocp4

import (
	"github.com/fusor/cpma/ocp4/oauth"
	"github.com/fusor/cpma/ocp4/secrets"
	configv1 "github.com/openshift/api/legacyconfig/v1"
	"github.com/sirupsen/logrus"
)

type Cluster struct {
	Master Master
}

type Master struct {
	OAuth   oauth.OAuthCRD
	Secrets []secrets.Secret
}

// Manifest holds a CRD object
type Manifest struct {
	Name string
	CRD  []byte
}

const OCP4InstallMsg = `To install OCP4 run the installer as follow in order to add CRDs:
' /openshift-install --dir $INSTALL_DIR create install-config'
'./openshift-install --dir install create manifests'
# Copy generated CRD manifest files  to '$INSTALL_DIR/openshift/'
# Edit them if needed, then run installation:
'./openshift-install --dir $INSTALL_DIR  create cluster'`

func (cluster *Cluster) Translate(masterconfig configv1.MasterConfig) {
	oauth, secrets, err := oauth.Translate(masterconfig.OAuthConfig)
	if err != nil {
		logrus.WithError(err).Fatalf("Unable to generate OAuth CRD from %+v", masterconfig.OAuthConfig)
	}
	cluster.Master.OAuth = *oauth
	cluster.Master.Secrets = secrets
}

// GenYAML returns the list of translated CRDs
func (cluster *Cluster) GenYAML() []Manifest {
	var manifests []Manifest
	manifest := Manifest{Name: "100_CPMA-cluster-config-oauth.yaml", CRD: cluster.Master.OAuth.GenYAML()}
	manifests = append(manifests, manifest)

	for _, secret := range cluster.Master.Secrets {
		filename := "100_CPMA-cluster-config-secret-" + secret.Metadata.Name + ".yaml"
		m := Manifest{Name: filename, CRD: secret.GenYAML()}
		manifests = append(manifests, m)
	}
	return manifests
}

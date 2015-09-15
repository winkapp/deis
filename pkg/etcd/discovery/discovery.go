/* Package discovery contains utlities for Etcd discovery. */
package discovery

import (
	"bytes"
	"io/ioutil"
)

const TokenFile = "/var/run/secrets/deis/etcd/discovery/token"
const ClusterDiscoveryURL = "http://%s:%s/v2/keys/deis/discovery/%s"
const ClusterSizeKey = "deis/discovery/%s/_config/size"

// Token reads the discovery token from the TokenFile and returns it.
func Token() ([]byte, error) {
	data, err := ioutil.ReadFile(TokenFile)
	if err != nil {
		return data, err
	}
	data = bytes.TrimSpace(data)
	return data, nil
}

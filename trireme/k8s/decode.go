package k8s

import (
	"k8s.io/kubernetes/pkg/api/latest"
	"k8s.io/kubernetes/pkg/runtime"
	//"encoding/json"
	"io/ioutil"
)

func DecodeFile(path string) (runtime.Object, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return latest.Codec.Decode(data)
}

func Decode(data []byte) (runtime.Object, error) {
	return latest.Codec.Decode(data)
}

// Read a Kubernetes file and decode it into an object.
// FIXME: This is an ugly first pass that does not use the k8s API. It only
// handles JSON.
/*
func DecodeFile(path string, o interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, o)
}
*/

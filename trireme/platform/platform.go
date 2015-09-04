/* Package platform provides Deis platform components.
 */
package platform

import (
	//"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/deis/deis/trireme/k8s"
	"github.com/deis/deis/trireme/storage"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/latest"
)

// Component describes a component of the Deis platform.
type Component struct {
	Name, Description                                 string
	RCs, Pods, Services, Namespaces, Volumes, Secrets []string
	Optional                                          bool
}

// InstallPrereqs installs Services, Namespaces, Volumes, and Secrets.
//
// This installs things that are considered prerequisites for the RC or Pod to
// function.
func (c *Component) InstallPrereqs(dir string) error {
	for _, ns := range c.Namespaces {
		p := filepath.Join(dir, ns)
		fmt.Printf("Installed namespace %s\n", p)
		if err := k8s.KubectlCreate(p); err != nil {
			return err
		}
	}
	for _, s := range c.Services {
		p := filepath.Join(dir, s)
		fmt.Printf("Installed services %s\n", p)
		if err := k8s.KubectlCreate(p); err != nil {
			return err
		}
	}
	for _, vol := range c.Volumes {
		p := filepath.Join(dir, vol)
		fmt.Printf("Installed volume %s\n", p)
		if err := k8s.KubectlCreate(p); err != nil {
			return err
		}
	}
	for _, sec := range c.Secrets {
		p := filepath.Join(dir, sec)
		fmt.Printf("Installed secret %s\n", p)
		if err := k8s.KubectlCreate(p); err != nil {
			return err
		}
	}
	return nil
}

// DeletePrereqs deletes a component's prerequisites.
//
// The order is Secrets, Volumes, Services, Namespaces.
func (c *Component) DeletePrereqs(dir string) error {
	for _, x := range c.Secrets {
		p := filepath.Join(dir, x)
		fmt.Printf("Removing Secret %s\n", p)
		if err := k8s.KubectlDelete(p); err != nil {
			return err
		}
	}
	for _, x := range c.Volumes {
		p := filepath.Join(dir, x)
		fmt.Printf("Removing Volume %s\n", p)
		if err := k8s.KubectlDelete(p); err != nil {
			return err
		}
	}
	for _, x := range c.Services {
		p := filepath.Join(dir, x)
		fmt.Printf("Removing Service %s\n", p)
		if err := k8s.KubectlDelete(p); err != nil {
			return err
		}
	}
	for _, x := range c.Namespaces {
		p := filepath.Join(dir, x)
		fmt.Printf("Removing Namespace %s\n", p)
		if err := k8s.KubectlDelete(p); err != nil {
			return err
		}
	}
	return nil
}

// Install installs pods and RCs (in that order).
func (c *Component) Install(dir string) error {
	for _, pod := range c.Pods {
		p := filepath.Join(dir, pod)
		fmt.Printf("Installing pod %s\n", p)
		if err := k8s.KubectlCreate(p); err != nil {
			return err
		}
	}
	for _, rc := range c.RCs {
		p := filepath.Join(dir, rc)
		fmt.Printf("Installing replication controller %s\n", p)
		if err := k8s.KubectlCreate(p); err != nil {
			return err
		}
	}
	return nil
}

// Delete removes a component's pods and rcs.
//
// Pods are deleted first, then RCs.
func (c *Component) Delete(dir string) error {
	for _, pod := range c.Pods {
		p := filepath.Join(dir, pod)
		fmt.Printf("Removing Pod %s\n", p)
		if err := k8s.KubectlDelete(p); err != nil {
			return err
		}
	}
	for _, rc := range c.RCs {
		p := filepath.Join(dir, rc)
		fmt.Printf("Removing RC %s\n", p)
		if err := k8s.KubectlDelete(p); err != nil {
			return err
		}
	}
	return nil
}

// InstallAll installs all the components in the given list.
//
// If optional is true, this will install packages marked optional. Otherwise,
// it will only install non-optional components.
func InstallAll(list []*Component, dir string, optional bool) error {
	for _, item := range list {
		if (optional && item.Optional) || !item.Optional {
			if err := item.InstallPrereqs(dir); err != nil {
				return err
			}
		}
	}
	for _, item := range list {
		if (optional && item.Optional) || !item.Optional {
			if err := item.Install(dir); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteAll removes the entire platform.
func DeleteAll(list []*Component, dir string) error {
	var incomplete bool
	for _, item := range list {
		if err := item.Delete(dir); err != nil {
			fmt.Printf("Error: %s", err)
			incomplete = true
		}
	}
	for _, item := range list {
		if err := item.DeletePrereqs(dir); err != nil {
			fmt.Printf("Error: %s", err)
			incomplete = true
		}
	}
	if incomplete {
		return errors.New("Incomplete deletion")
	}
	return nil
}

// filterFunc takes data, a component , and storage and filters the data.
type filterFunc func([]byte, *Component, storage.Storer) ([]byte, error)

var filterMap = map[string]filterFunc{
	"RCs": rcFilter,
}

func rcFilter(data []byte, comp *Component, store storage.Storer) ([]byte, error) {
	var rc *api.ReplicationController
	if o, err := k8s.Decode(data); err != nil {
		return data, err
	} else {
		rc = o.(*api.ReplicationController)
	}

	rc.APIVersion = "v1"
	rc.Kind = "ReplicationController"

	fmt.Printf("APIVersion:%s, Kind: %s\n", rc.APIVersion, rc.Kind)

	kname := rc.Name

	val, err := store.Get(kname, "image")
	if err == nil && len(val) > 0 {
		orig := rc.Spec.Template.Spec.Containers[0].Image
		rc.Spec.Template.Spec.Containers[0].Image = val
		fmt.Printf("===> Replacing %s with %s\n", orig, val)
	}

	return latest.Codec.Encode(rc)
	//return json.MarshalIndent(rc, "", "  ")
}

// RebuildDefs reads a set of files, rebuilds, and writes them.
//
// To rebuild, this applies a finite set of hard-coded rules to transform
// a source file via configuration directives into a destination.
func RebuildDefs(src, dest string, comps []*Component, store storage.Storer) error {

	rebuild := []string{"Services", "Namespaces", "Volumes", "Secrets", "RCs", "Pods"}
	for _, comp := range comps {
		cv := reflect.Indirect(reflect.ValueOf(comp))
		for _, n := range rebuild {
			fv := cv.FieldByName(n)
			if fv.Len() > 0 {
				for _, file := range fv.Interface().([]string) {

					src := filepath.Join(src, file)
					dest := filepath.Join(dest, file)

					fmt.Printf("Rebuilding file %s to %s\n", src, dest)
					// Ensure that the basedir exists in dest
					os.MkdirAll(filepath.Dir(dest), 0755)

					in, err := ioutil.ReadFile(src)
					if err != nil {
						return err
					}

					if fn, ok := filterMap[n]; ok {
						fmt.Printf("Found filter func for %s\n", n)
						in, err = fn(in, comp, store)
						if err != nil {
							return err
						}
					}
					if err := ioutil.WriteFile(dest, in, 0755); err != nil {
						return err
					}

				}
			}
		}
	}
	return nil
}

/* Package platform provides Deis platform components.
 */
package platform

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/deis/deis/trireme/k8s"
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

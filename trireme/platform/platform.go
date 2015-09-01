/* Package platform provides Deis platform components.
 */
package platform

import (
	"fmt"
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
func (c *Component) InstallPrereqs() error {
	for _, ns := range c.Namespaces {
		fmt.Printf("Installed namespace %s\n", ns)
	}
	for _, s := range c.Services {
		fmt.Printf("Installed services %s\n", s)
	}
	for _, vol := range c.Volumes {
		fmt.Printf("Installed volume %s\n", vol)
	}
	for _, sec := range c.Secrets {
		fmt.Printf("Installed secret %s\n", sec)
	}
	return nil
}

// Install installs pods and RCs (in that order).
func (c *Component) Install() error {
	for _, pod := range c.Pods {
		fmt.Printf("Installing pod %s\n", pod)
	}
	for _, rc := range c.RCs {
		fmt.Printf("Installing replication controller %s\n", rc)
	}
	return nil
}

// InstallAll installs all the components in the given list.
//
// If optional is true, this will install packages marked optional. Otherwise,
// it will only install non-optional components.
func InstallAll(list []*Component, optional bool) error {
	for _, item := range list {
		if (optional && item.Optional) || !item.Optional {
			if err := item.InstallPrereqs(); err != nil {
				return err
			}
		}
	}
	for _, item := range list {
		if (optional && item.Optional) || !item.Optional {
			if err := item.Install(); err != nil {
				return err
			}
		}
	}
	return nil
}

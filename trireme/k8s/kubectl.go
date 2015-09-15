package k8s

import (
	"fmt"
	"os/exec"
)

// The plan is to eventually call all of these commands directly from the
// k8s API. For the sake of expedience, this is a convenience wrapper until
// we get there.

// KubectlCreate calls 'kubectl create'.
func KubectlCreate(file string) error {
	out, err := exec.Command("kubectl", "create", "-f", file).CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", out)
		return err
	}
	fmt.Printf("Scheduled %s", out)
	return nil
}

func KubectlDelete(filename string) error {
	out, err := exec.Command("kubectl", "delete", "-f", filename).CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", out)
		return err
	}
	fmt.Printf("Deleting %s", out)
	return nil
}

// +build framework

package framework

import (
	"fmt"
	"os/exec"
	"time"
)

// KubectlBuilder is used to build, customize and execute a kubectl Command.
// Add more functions to customize the builder as needed.
type KubectlBuilder struct {
	cmd     *exec.Cmd
	timeout <-chan time.Time
}

// LookForStringInPodExec looks for the given string in the output of a command
// executed in a specific pod container.
func LookForStringInPodExec(ns, podName string, command []string, expectedString string, timeout time.Duration) (result string, err error) {
	return lookForString(expectedString, timeout, func() string {
		// use the first container
		args := []string{"exec", podName, fmt.Sprintf("--namespace=%v", ns), "--"}
		args = append(args, command...)
		return RunKubectlOrDie(ns, args...)
	})
}

// lookForString looks for the given string in the output of fn, repeatedly calling fn until
// the timeout is reached or the string is found. Returns last log and possibly
// error if the string was not found.
func lookForString(expectedString string, timeout time.Duration, fn func() string) (result string, err error) {
	for t := time.Now(); time.Since(t) < timeout; time.Sleep(Poll) {
		result = fn()
		if strings.Contains(result, expectedString) {
			return
		}
	}
	err = fmt.Errorf("Failed to find \"%s\", last result: \"%s\"", expectedString, result)
	return
}

// RunKubectlOrDie is a convenience wrapper over kubectlBuilder
func RunKubectlOrDie(namespace string, args ...string) string {
	return NewKubectlCommand(namespace, args...).ExecOrDie(namespace)
}

// NewKubectlCommand returns a KubectlBuilder for running kubectl.
func NewKubectlCommand(namespace string, args ...string) *KubectlBuilder {
	b := new(KubectlBuilder)
	tk := e2ekubectl.NewTestKubeconfig(TestContext.CertDir, TestContext.Host, TestContext.KubeConfig, TestContext.KubeContext, TestContext.KubectlPath, namespace)
	b.cmd = tk.KubectlCmd(args...)
	return b
}

// LookForStringInPodExec looks for the given string in the output of a command
// executed in a specific pod container.
func LookForStringInPodExec(ns, podName string, command []string, expectedString string, timeout time.Duration) (result string, err error) {
	return lookForString(expectedString, timeout, func() string {
		// use the first container
		args := []string{"exec", podName, fmt.Sprintf("--namespace=%v", ns), "--"}
		args = append(args, command...)
		return RunKubectlOrDie(ns, args...)
	})
}

// lookForString looks for the given string in the output of fn, repeatedly calling fn until
// the timeout is reached or the string is found. Returns last log and possibly
// error if the string was not found.
func lookForString(expectedString string, timeout time.Duration, fn func() string) (result string, err error) {
	for t := time.Now(); time.Since(t) < timeout; time.Sleep(Poll) {
		result = fn()
		if strings.Contains(result, expectedString) {
			return
		}
	}
	err = fmt.Errorf("Failed to find \"%s\", last result: \"%s\"", expectedString, result)
	return
}

// RunKubectlOrDie is a convenience wrapper over kubectlBuilder
func RunKubectlOrDie(namespace string, args ...string) string {
	return NewKubectlCommand(namespace, args...).ExecOrDie(namespace)
}

// RunKubectl is a convenience wrapper over kubectlBuilder
func RunKubectl(namespace string, args ...string) (string, error) {
	return NewKubectlCommand(namespace, args...).Exec()
}

// RunKubectlOrDieInput is a convenience wrapper over kubectlBuilder that takes input to stdin
func RunKubectlOrDieInput(namespace string, data string, args ...string) string {
	return NewKubectlCommand(namespace, args...).WithStdinData(data).ExecOrDie(namespace)
}

// RunKubectlInput is a convenience wrapper over kubectlBuilder that takes input to stdin
func RunKubectlInput(namespace string, data string, args ...string) (string, error) {
	return NewKubectlCommand(namespace, args...).WithStdinData(data).Exec()
}

// NewKubectlCommand returns a KubectlBuilder for running kubectl.
func NewKubectlCommand(namespace string, args ...string) *KubectlBuilder {
	b := new(KubectlBuilder)
	tk := e2ekubectl.NewTestKubeconfig(TestContext.CertDir, TestContext.Host, TestContext.KubeConfig, TestContext.KubeContext, TestContext.KubectlPath, namespace)
	b.cmd = tk.KubectlCmd(args...)
	return b
}

func isTimeout(err error) bool {
	switch err := err.(type) {
	case *url.Error:
		if err, ok := err.Err.(net.Error); ok && err.Timeout() {
			return true
		}
	case net.Error:
		if err.Timeout() {
			return true
		}
	}
	return false
}

// ExecOrDie runs the kubectl executable or dies if error occurs.
func (b KubectlBuilder) ExecOrDie(namespace string) string {
	str, err := b.Exec()
	// In case of i/o timeout error, try talking to the apiserver again after 2s before dying.
	// Note that we're still dying after retrying so that we can get visibility to triage it further.
	if isTimeout(err) {
		Logf("Hit i/o timeout error, talking to the server 2s later to see if it's temporary.")
		time.Sleep(2 * time.Second)
		retryStr, retryErr := RunKubectl(namespace, "version")
		Logf("stdout: %q", retryStr)
		Logf("err: %v", retryErr)
	}
	ExpectNoError(err)
	return str
}

// ExpectNoError checks if "err" is set, and if so, fails assertion while logging the error.
func ExpectNoError(err error, explain ...interface{}) {
	ExpectNoErrorWithOffset(1, err, explain...)
}

// ExpectNoErrorWithOffset checks if "err" is set, and if so, fails assertion while logging the error at "offset" levels above its caller
// (for example, for call chain f -> g -> ExpectNoErrorWithOffset(1, ...) error would be logged for "f").
func ExpectNoErrorWithOffset(offset int, err error, explain ...interface{}) {
	gomega.ExpectWithOffset(1+offset, err).NotTo(gomega.HaveOccurred(), explain...)
}

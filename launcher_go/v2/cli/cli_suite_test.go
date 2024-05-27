package cli_test

import (
	"testing"

	"github.com/discourse/discourse_docker/launcher_go/v2/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cli Suite")
}

var _ = BeforeSuite(func() {
	utils.CommitWait = 0
})

package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"bytes"
	"context"
	ddocker "github.com/discourse_docker/go"
	"io"
	"os"
	"os/exec"
	"strings"
)

type FakeCmdRunner struct {
	Cmd      *exec.Cmd
	RunCalls int
}

func (r *FakeCmdRunner) Run() error {
	r.RunCalls = r.RunCalls + 1
	return nil
}

// Swap out CmdRunner with a fake instance that also returns created ICmdRunners on a channel
// so tests can inspect commands the moment they're run
func CreateNewFakeCmdRunner(c chan ddocker.ICmdRunner) func(cmd *exec.Cmd) ddocker.ICmdRunner {
	return func(cmd *exec.Cmd) ddocker.ICmdRunner {
		cmdRunner := &FakeCmdRunner{Cmd: cmd, RunCalls: 0}
		c <- cmdRunner
		return cmdRunner
	}
}

var _ = Describe("Main", func() {
	var testDir string
	var out *bytes.Buffer
	testConfDir := "./test/containers"
	testTemplatesDir := "./test"
	var runArgs ddocker.RunArgs

	BeforeEach(func() {
		out = &bytes.Buffer{}
		ddocker.Out = out
		testDir, _ = os.MkdirTemp("", "ddocker-test")

		runArgs = ddocker.NewRunArgs(context.Background())
		runArgs.ConfDir = testConfDir
		runArgs.TemplatesDir = testTemplatesDir
		runArgs.OutputDir = testDir
	})
	AfterEach(func() {
		os.RemoveAll(testDir)
	})

	It("should allow concatenated templates", func() {
		runner := ddocker.RawYamlCmd{Config: "test"}
		runner.Run(&runArgs)
		Expect(out.String()).To(ContainSubstring("DISCOURSE_DEVELOPER_EMAILS: 'me@example.com,you@example.com'"))
		Expect(out.String()).To(ContainSubstring("_FILE_SEPERATOR_"))
		Expect(out.String()).To(ContainSubstring("version: tests-passed"))
	})

	It("should output docker compose cmd to config name's subdir", func() {
		runner := ddocker.DockerComposeCmd{Config: "test"}
		err := runner.Run(&runArgs)
		Expect(err).To(BeNil())
		out, err := os.ReadFile(testDir + "/test/test.config.yaml")
		Expect(err).To(BeNil())
		Expect(string(out[:])).To(ContainSubstring("DISCOURSE_DEVELOPER_EMAILS: 'me@example.com,you@example.com'"))
	})

	It("should clean after the command", func() {
		runner := ddocker.DockerComposeCmd{Config: "test"}
		runner.Run(&runArgs)
		runner2 := ddocker.CleanCmd{Config: "test"}
		err := runner2.Run(&runArgs)
		Expect(err).To(BeNil())
		_, err = os.ReadFile(testDir + "/test/test.config.yaml")
		Expect(err).ToNot(BeNil())
	})

	Context("When running docker commands", func() {

		var CmdCreatorWatcher chan ddocker.ICmdRunner

		BeforeEach(func() {
			CmdCreatorWatcher = make(chan ddocker.ICmdRunner)
			ddocker.CmdRunner = CreateNewFakeCmdRunner(CmdCreatorWatcher)
		})
		AfterEach(func() {
			close(CmdCreatorWatcher)
		})

		It("Should run docker build with correct arguments", func() {
			runner := ddocker.DockerBuildCmd{Config: "test"}
			go runner.Run(&runArgs)
			icmd := <-CmdCreatorWatcher
			cmd, _ := icmd.(*FakeCmdRunner)
			Expect(cmd.RunCalls).To(Equal(1))
			Expect(cmd.Cmd.String()).To(ContainSubstring("docker build"))
			Expect(cmd.Cmd.String()).To(ContainSubstring("--build-arg DISCOURSE_DEVELOPER_EMAILS"))
			Expect(cmd.Cmd.Dir).To(Equal(testDir + "/test"))
			Expect(cmd.Cmd.Env).To(ContainElement("DISCOURSE_DB_PASSWORD=SOME_SECRET"))
			buf := new(strings.Builder)
			io.Copy(buf, cmd.Cmd.Stdin)
			// docker build's stdin is a dockerfile
			Expect(buf.String()).To(ContainSubstring("COPY ./test.config.yaml /temp-config.yaml"))
			Expect(buf.String()).To(ContainSubstring("--skip-tags=precompile,migrate,db"))
		})

		It("Should run docker migrate with correct arguments", func() {
			runner := ddocker.DockerMigrateCmd{Config: "test"}
			go runner.Run(&runArgs)
			icmd := <-CmdCreatorWatcher
			cmd, _ := icmd.(*FakeCmdRunner)
			Expect(cmd.RunCalls).To(Equal(1))
			Expect(cmd.Cmd.String()).To(ContainSubstring("docker run"))
			Expect(cmd.Cmd.String()).To(ContainSubstring("-e DISCOURSE_DEVELOPER_EMAILS"))
			Expect(cmd.Cmd.Env).To(ContainElement("DISCOURSE_DB_PASSWORD=SOME_SECRET"))
			buf := new(strings.Builder)
			io.Copy(buf, cmd.Cmd.Stdin)
			// docker run's stdin is a pups config
			Expect(buf.String()).To(ContainSubstring("path: /etc/service/nginx/run"))
		})

		It("Should run docker run followed by docker commit when configuring", func() {
			runner := ddocker.DockerConfigureCmd{Config: "test"}
			go runner.Run(&runArgs)
			icmd := <-CmdCreatorWatcher
			cmd, _ := icmd.(*FakeCmdRunner)
			Expect(cmd.RunCalls).To(Equal(1))
			Expect(cmd.Cmd.String()).To(ContainSubstring("docker run"))
			Expect(cmd.Cmd.String()).To(ContainSubstring("-e DISCOURSE_DEVELOPER_EMAILS"))
			Expect(cmd.Cmd.Env).To(ContainElement("DISCOURSE_DB_PASSWORD=SOME_SECRET"))
			buf := new(strings.Builder)
			io.Copy(buf, cmd.Cmd.Stdin)
			// docker run's stdin is a pups config
			Expect(buf.String()).To(ContainSubstring("path: /etc/service/nginx/run"))

			icmd = <-CmdCreatorWatcher
			cmd, _ = icmd.(*FakeCmdRunner)
			Expect(cmd.RunCalls).To(Equal(1))
			Expect(cmd.Cmd.String()).To(ContainSubstring("docker commit"))
			Expect(cmd.Cmd.String()).To(ContainSubstring("discourse-build"))
			Expect(cmd.Cmd.String()).To(ContainSubstring("local_discourse/test"))
			Expect(cmd.Cmd.Env).ToNot(ContainElement("DISCOURSE_DB_PASSWORD=SOME_SECRET"))
		})
	})

})

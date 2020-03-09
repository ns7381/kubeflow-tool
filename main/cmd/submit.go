package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"regexp"
)
var (
	envs        []string
	selectors   []string
	configFiles []string
	tolerations []string
	dataset     []string
	dataDirs    []string
	annotations []string
)

var (
	submitLong = `Submit a job.

Available Commands:
  tfjob,tf             Submit a TFJob.
  horovod,hj           Submit a Horovod Job.
  mpijob,mpi           Submit a MPIJob.
  standalonejob,sj     Submit a standalone Job.
  tfserving,tfserving  Submit a Serving Job.
  volcanojob,vj        Submit a VolcanoJob.
    `
)
// Job Max lenth should be 49
const JobMaxLength int = 49
const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

func NewSubmitCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "submit",
		Short: "Submit a job.",
		Long:  submitLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(NewSubmitTFJobCommand())

	return command
}
type submitArgs struct {
	Image       string                               `yaml:"image"`       // --image
	GPUCount    int                                  `yaml:"gpuCount"`    // --gpuCount
	WorkingDir  string                               `yaml:"workingDir"`  // --workingDir
	Envs        map[string]string                    `yaml:"envs"`        // --envs
	Command     string                               `yaml:"command"`
	// for horovod
	WorkerCount int    `yaml:"workers"` // --workers
}

func (s submitArgs) check() error {
	if name == "" {
		return fmt.Errorf("--name must be set")
	}

	// return fmt.Errorf("must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.")
	err := ValidateJobName(name)
	if err != nil {
		return err
	}

	return nil
}
func ValidateJobName(value string) error {
	if len(value) > JobMaxLength {
		return fmt.Errorf("The len %d of name %s is too long, it should be less than %d",
			len(value),
			value,
			JobMaxLength)
	}
	if !dns1123LabelRegexp.MatchString(value) {
		return fmt.Errorf("The job name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.")
	}
	return nil
}
func (submitArgs *submitArgs) addCommonFlags(command *cobra.Command) {

	// create subcommands
	command.Flags().StringVar(&name, "name", "", "override name")
	_ = command.MarkFlagRequired("name")
	command.Flags().StringVar(&submitArgs.Image, "image", config.DefaultImage, "the docker image name of training job")
	// command.MarkFlagRequired("image")
	command.Flags().IntVar(&submitArgs.GPUCount, "gpus", 0,
		"the GPU count of each worker to run the training.")
	// command.Flags().StringVar(&submitArgs.DataDir, "dataDir", "", "the data dir. If you specify /data, it means mounting hostpath /data into container path /data")
	command.Flags().IntVar(&submitArgs.WorkerCount, "workers", 1,
		"the worker number to run the distributed training.")
	// command.MarkFlagRequired("syncSource")
	command.Flags().StringVar(&submitArgs.WorkingDir, "working-dir", "/", "working directory to extract the code. If using syncMode, the $workingDir/code contains the code")
	command.Flags().StringArrayVarP(&envs, "env", "e", []string{}, "the environment variables")
}
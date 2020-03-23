package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"os"
	"strings"
)


func NewSubmitTFJobCommand() *cobra.Command {
	var (
		submitArgs submitTFJobArgs
	)


	var command = &cobra.Command{
		Use:     "tfjob",
		Short:   "Submit TFJob as training job.",
		Long:   "Submit TFJob as training job. Running container mount your netdisk to /notebook",
		Aliases: []string{"tf"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			client := initKubeClient()

			err := submitTFJob(client, args, &submitArgs)
			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	submitArgs.addCommonFlags(command)

	// TFJob
	command.Flags().StringVar(&submitArgs.WorkerImage, "worker-image", config.DefaultImage, "the docker image for tensorflow workers")

	command.Flags().StringVar(&submitArgs.PSImage, "ps-image", config.DefaultImage, "the docker image for tensorflow workers")

	command.Flags().IntVar(&submitArgs.PSCount, "ps", 0, "the number of the parameter servers.")

	command.Flags().StringVar(&submitArgs.WorkerCpu, "worker-cpu", "1", "the cpu resource to use for the worker, like 1 for 1 core.")
	command.Flags().StringVar(&submitArgs.WorkerMemory, "worker-memory", "1Gi", "the memory resource to use for the worker, like 1Gi.")
	command.Flags().StringVar(&submitArgs.PSCpu, "ps-cpu", "1", "the cpu resource to use for the parameter servers, like 1 for 1 core.")
	command.Flags().StringVar(&submitArgs.PSMemory, "ps-memory", "1Gi", "the memory resource to use for the parameter servers, like 1Gi.")
	// Estimator
	command.Flags().BoolVar(&submitArgs.UseChief, "chief", false, "enable chief, which is required for estimator.")
	command.Flags().BoolVar(&submitArgs.UseEvaluator, "evaluator", false, "enable evaluator, which is optional for estimator.")
	command.Flags().StringVar(&submitArgs.ChiefCpu, "chief-cpu", "", "the cpu resource to use for the Chief, like 1 for 1 core.")
	command.Flags().StringVar(&submitArgs.ChiefMemory, "chief-memory", "", "the memory resource to use for the Chief, like 1Gi.")
	command.Flags().StringVar(&submitArgs.EvaluatorCpu, "evaluator-cpu", "", "the cpu resource to use for the evaluator, like 1 for 1 core.")
	command.Flags().StringVar(&submitArgs.EvaluatorMemory, "evaluator-memory", "", "the memory resource to use for the evaluator, like 1Gi.")

	return command
}

func submitTFJob(client dynamic.Interface, args []string, submitArgs *submitTFJobArgs) (err error) {
	err = submitArgs.prepare(args)
	if err != nil {
		return err
	}
	tfJobRes := schema.GroupVersionResource{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}
	tfJob := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "TFJob",
			"metadata": map[string]interface{}{
				"name": name,
				"namespace": config.Namespace,
				"labels": map[string]interface{}{
					"createdBy": loginUser,
				},
			},
			"spec": map[string]interface{}{
				"tfReplicaSpecs": map[string]interface{}{
					"PS": map[string]interface{}{
						"replicas":      submitArgs.PSCount,
						"restartPolicy": "Never",
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []map[string]interface{}{
									{
										"name":  "tensorflow",
										"image": submitArgs.PSImage,
										"workingDir": submitArgs.WorkingDir,
										"command": []string{"sh", "-c", submitArgs.Command},
										"resources": map[string]interface{}{
											"limits": map[string]interface{}{
												"cpu": submitArgs.PSCpu,
												"memory": submitArgs.PSMemory,
												"nvidia.com/gpu": "0",
											},
											"requests": map[string]interface{}{
												"cpu": submitArgs.PSCpu,
												"memory": submitArgs.PSMemory,
												"nvidia.com/gpu": "0",
											},
										},
										"volumeMounts": []map[string]interface{}{
											{
												"name": "netdisk",
												"mountPath": "/notebook",
											},
										},
									},
								},
								"volumes": []map[string]interface{}{
									{
										"name": "netdisk",
										"persistentVolumeClaim": map[string]interface{}{
											"claimName": "claim-" + loginUser,
										},
									},
								},
							},
						},
					},
					"Worker": map[string]interface{}{
						"replicas":      submitArgs.WorkerCount,
						"restartPolicy": "Never",
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"containers": []map[string]interface{}{
									{
										"name":  "tensorflow",
										"image": submitArgs.WorkerImage,
										"workingDir": submitArgs.WorkingDir,
										"command": []string{"sh", "-c", submitArgs.Command},
										"resources": map[string]interface{}{
											"limits": map[string]interface{}{
												"cpu": submitArgs.WorkerCpu,
												"memory": submitArgs.WorkerMemory,
												"nvidia.com/gpu": submitArgs.GPUCount,
											},
											"requests": map[string]interface{}{
												"cpu": submitArgs.WorkerCpu,
												"memory": submitArgs.WorkerMemory,
												"nvidia.com/gpu": submitArgs.GPUCount,
											},
										},
										"volumeMounts": []map[string]interface{}{
											{
												"name": "netdisk",
												"mountPath": "/notebook",
											},
										},
									},
								},
								"volumes": []map[string]interface{}{
									{
										"name": "netdisk",
										"persistentVolumeClaim": map[string]interface{}{
											"claimName": "claim-" + loginUser,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating tfJob...")
	result, err := client.Resource(tfJobRes).Namespace(config.Namespace).Create(tfJob, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created tfJob %q.\n", result.GetName())
	return err
}

type submitTFJobArgs struct {
	WorkerImage     string                       `yaml:"workerImage"` // --workerImage
	PSCount   int    `yaml:"ps"`        // --ps
	PSImage   string `yaml:"psImage"`   // --psImage
	WorkerCpu string `yaml:"workerCPU"` // --workerCpu
	//WorkerNodeSelectors map[string]string `yaml:"workerNodeSelectors"` // --worker-selector
	WorkerMemory   string `yaml:"workerMemory"`   // --workerMemory
	PSCpu          string `yaml:"psCPU"`          // --psCpu
	PSMemory       string `yaml:"psMemory"`       // --psMemory
	// For esitmator, it reuses workerImage
	UseChief     bool `yaml:",omitempty"` // --chief
	ChiefCount   int  `yaml:"chief"`
	UseEvaluator bool `yaml:",omitempty"` // --evaluator
	ChiefCpu     string `yaml:"chiefCPU"`     // --chiefCpu
	ChiefMemory  string `yaml:"chiefMemory"`  // --chiefMemory
	EvaluatorCpu string `yaml:"evaluatorCPU"` // --evaluatorCpu
	//EvaluatorNodeSelectors map[string]string `yaml:"evaluatorNodeSelectors"` // --evaluator-selector
	EvaluatorMemory string `yaml:"evaluatorMemory"` // --evaluatorMemory
	EvaluatorCount  int    `yaml:"evaluator"`

	// for common args
	submitArgs `yaml:",inline"`
}

func (submitArgs submitTFJobArgs) check() error {
	err := submitArgs.submitArgs.check()
	if err != nil {
		return err
	}


	if submitArgs.WorkerCount == 0 && !submitArgs.UseChief {
		return fmt.Errorf("--workers must be greater than 0 in distributed training")
	}

	return nil
}
func (submitArgs *submitTFJobArgs) prepare(args []string) (err error) {
	submitArgs.Command = strings.Join(args, " ")

	// 1. Use specified runtime to transform
	err = submitArgs.check()
	if err != nil {
		return err
	}
	return nil
}

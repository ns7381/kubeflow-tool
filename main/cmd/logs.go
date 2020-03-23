// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/ns7381/kubeflow-tool/main/podlogs"
	"os"
	"time"

	tlogs "github.com/ns7381/kubeflow-tool/main/printer/base/logs"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewLogsCommand() *cobra.Command {
	var outerArgs = &podlogs.OuterRequestArgs{}
	var command = &cobra.Command{
		Use:   "logs training job",
		Short: "print the logs for a task of the training job",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name = args[0]
			outerArgs.KubeClient = initKubeClientset()

			//tfJobRes := schema.GroupVersionResource{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}
			//job, err := client.Resource(tfJobRes).Namespace(config.Namespace).Get(name, metav1.GetOptions{})
			//if err != nil {
			//	panic(err)
			//}

			outerArgs.Namespace = config.Namespace
			outerArgs.RetryCount = 5
			outerArgs.RetryTimeout = time.Millisecond
			names := []string{name + "-worker-0"}
			logPrinter, err := tlogs.NewPodLogPrinter(names, outerArgs)
			if err != nil {
				log.Errorf(err.Error())
				os.Exit(1)
			}
			code, err := logPrinter.Print()
			if err != nil {
				log.Errorf("%s, %s", err.Error(), "please use \"arena get\" to get more information.")
				os.Exit(1)
			} else if code != 0 {
				os.Exit(code)
			}
		},
	}

	command.Flags().BoolVarP(&outerArgs.Follow, "follow", "f", false, "Specify if the logs should be streamed.")
	command.Flags().StringVar(&outerArgs.SinceSeconds, "since", "", "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().StringVar(&outerArgs.SinceTime, "since-time", "", "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().IntVarP(&outerArgs.Tail, "tail", "t", -1, "Lines of recent log file to display. Defaults to -1 with no selector, showing all log lines otherwise 10, if a selector is provided.")
	command.Flags().BoolVar(&outerArgs.Timestamps, "timestamps", false, "Include timestamps on each line in the log output")
	command.Flags().StringVarP(&outerArgs.PodName, "instance", "i", "", "Specify the task instance to get log")
	return command
}

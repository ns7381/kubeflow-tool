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
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"strings"
	"text/tabwriter"

	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "list all the training jobs",
		Run: func(cmd *cobra.Command, args []string) {
			client := initKubeClient()

			tfJobRes := schema.GroupVersionResource{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}
			labelSelector := fmt.Sprintf("createdBy=%s", loginUser)
			result, err := client.Resource(tfJobRes).Namespace(config.Namespace).List(metav1.ListOptions{LabelSelector: labelSelector,})
			if err != nil {
				panic(err)
			}
			displayTrainingJobList(result, false)
		},
	}

	return command
}

func displayTrainingJobList(jobInfoList *unstructured.UnstructuredList, displayGPU bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	labelField := []string{"NAME", "STATE", "CREATE_TIME"}

	PrintLine(w, labelField...)

	for _, jobInfo := range jobInfoList.Items {
		conditions := jobInfo.Object["status"].(map[string]interface{})["conditions"].([]interface{})
		condition := conditions[len(conditions) - 1]
		PrintLine(w, jobInfo.GetName(),
			condition.(map[string]interface{})["type"].(string),
			jobInfo.GetCreationTimestamp().String(),
			)
	}
	_ = w.Flush()
}

func PrintLine(w io.Writer, fields ...string) {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	buffer := strings.Join(fields, "\t")
	_, _ = fmt.Fprintln(w, buffer)
}

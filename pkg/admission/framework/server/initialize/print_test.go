package initialize

import (
	"fmt"
	"testing"

	"k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestPrint(t *testing.T) {
	c := &HandlerConfig{
		Configs: []GroupVersionResource{
			{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
				Handlers: map[v1beta1.OperationType]*Handler{
					v1beta1.Create: {
						TimeoutSecond: 11,
						Processors: []Processor{
							{
								Module: "moduleA",
								Name:   "processorA",
							},
							{
								Module: "moduleB",
								Name:   "processorB",
							},
						},
					},
					v1beta1.Update: {
						TimeoutSecond: 11,
						Processors: []Processor{
							{
								Module: "moduleC",
								Name:   "processorC",
							},
							{
								Module: "moduleD",
								Name:   "processorD",
							},
						},
					},
				},
			},
			{
				GroupVersionResource: schema.GroupVersionResource{
					Group:    "apps",
					Version:  "v1",
					Resource: "statefulsets",
				},
				Handlers: map[v1beta1.OperationType]*Handler{
					v1beta1.Create: {
						TimeoutSecond: 11,
						Processors: []Processor{
							{
								Module: "moduleE",
								Name:   "processorE",
							},
							{
								Module: "moduleF",
								Name:   "processorF",
							},
						},
					},
					v1beta1.Update: {
						TimeoutSecond: 11,
						Processors: []Processor{
							{
								Module: "moduleG",
								Name:   "processorG",
							},
							{
								Module: "moduleH",
								Name:   "processorH",
							},
						},
					},
				},
			},
		},
	}
	fmt.Println(c.String())
}

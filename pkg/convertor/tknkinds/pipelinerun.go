package tknkinds

import (
	"encoding/json"
	"time"

	toto "github.com/in-toto/in-toto-golang/in_toto"
	v1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

//PipelinerunToIntotoProvenance :
func PipelinerunToIntotoProvenance(pipelinerun *v1beta1.PipelineRun) ([]byte, error) {
	currenttime := time.Now()
	currenttime = currenttime.Add(30 * 24 * time.Hour)

	var keys = make(map[string]toto.Key)

	var intotoSteps []toto.Step
	var expMats [][]string

	for _, taskrunStatus := range pipelinerun.Status.TaskRuns {

		params := getPipelineParams(pipelinerun)

		for _, stepState := range taskrunStatus.Status.Steps {

			intotoStep := toto.Step{
				Type:            "step",
				ExpectedCommand: GetExpectedCommand(getStep(stepState, taskrunStatus), params),
				SupplyChainItem: toto.SupplyChainItem{
					Name:              stepState.Name,
					ExpectedMaterials: append(expMats, GetExpectedMaterials(stepState)...),
					//ExpectedProducts:  getExpectedProducts(taskrun, false),
				},
			}
			intotoSteps = append(intotoSteps, intotoStep)
		}
	}

	var metablock = toto.Metablock{
		Signed: toto.Layout{
			Type:    pipelinerun.Name,
			Expires: currenttime.Format("2006-01-02T15:04:05Z"),
			Steps:   intotoSteps,
			Inspect: nil,
			Keys:    keys,
		},
	}

	return json.MarshalIndent(metablock, "", "    ")
}

// Get parameters from tasks that can be substituted into the commands and arguments
func getPipelineParams(pr *v1beta1.PipelineRun) map[string]string {

	// get parameters
	params := make(map[string]string)
	for _, p := range pr.Spec.Params {
		params[p.Name] = p.Value.StringVal
	}

	return params
}

func getStep(stepState v1beta1.StepState, trs *v1beta1.PipelineRunTaskRunStatus) v1beta1.Step {
	name := stepState.Name
	if trs.Status.TaskSpec != nil {
		for _, s := range trs.Status.TaskSpec.Steps {
			if s.Name == name {
				return s
			}
		}
	}

	return v1beta1.Step{}
}

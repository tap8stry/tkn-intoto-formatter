package parser

import (
	"fmt"

	v1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned/scheme"
)

// parseTknPipeline parses pipeline
func parseTknPipeline(r []byte) (*v1beta1.Pipeline, error) {
	var pipeline v1beta1.Pipeline
	if r == nil {
		return nil, nil
	}

	_, _, err := scheme.Codecs.UniversalDeserializer().Decode(r, nil, &pipeline)
	if err != nil {
		fmt.Printf("error parsing `pipeline' object: %v", err)
		return nil, err
	}

	return &pipeline, nil
}

// parseTknPipelineRun parses pipelinerun
func parseTknPipelineRun(r []byte) (*v1beta1.PipelineRun, error) {
	var pipelinerun v1beta1.PipelineRun
	if r == nil {
		return nil, nil
	}

	_, _, err := scheme.Codecs.UniversalDeserializer().Decode(r, nil, &pipelinerun)
	if err != nil {
		fmt.Printf("error parsing `pipelinerun' object: %v", err)
		return nil, err
	}

	return &pipelinerun, nil
}

// parseTknTask parses Task
func parseTknTask(r []byte) (*v1beta1.Task, error) {
	if r == nil {
		return nil, nil
	}
	var task v1beta1.Task
	_, _, err := scheme.Codecs.UniversalDeserializer().Decode(r, nil, &task)
	if err != nil {
		fmt.Printf("error parsing `task' object: %v", err)
		return nil, err
	}
	return &task, nil
}

// parseTknTaskRun parses Taskrun
func parseTknTaskRun(r []byte) (*v1beta1.TaskRun, error) {
	if r == nil {
		return nil, nil
	}
	var taskrun v1beta1.TaskRun
	_, _, err := scheme.Codecs.UniversalDeserializer().Decode(r, nil, &taskrun)
	if err != nil {
		fmt.Printf("error parsing `taskrun' object: %v", err)
		return nil, err
	}
	return &taskrun, nil
}

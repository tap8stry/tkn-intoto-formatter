package parser

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/tap8stry/tkn-intoto-formatter/pkg/common"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned/scheme"
)

const (
	task           string = "Task"
	taskrun        string = "TaskRun"
	pipeline       string = "Pipeline"
	pipelinerun    string = "PipelineRun"
	triggerbinding string = "TriggerBinding"
)

//GetTknResources :
func GetTknResources(file string) (common.TknResources, error) {
	acceptedK8sTypes := regexp.MustCompile(fmt.Sprintf("(%s|%s|%s|%s|%s)",
		task, pipeline, taskrun, pipelinerun, triggerbinding))
	var tknRes common.TknResources
	if common.IsYAMLFile(file) {
		if filebuf, err := ioutil.ReadFile(file); err == nil {
			fileAsString := string(filebuf[:])
			sepYamlfiles := strings.Split(fileAsString, "---")
			for _, f := range sepYamlfiles {
				if f == "\n" || f == "" {
					// ignore empty cases
					continue
				}
				decode := scheme.Codecs.UniversalDeserializer().Decode
				_, groupVersionKind, err := decode([]byte(f), nil, nil)
				if err != nil {
					continue
				}
				if acceptedK8sTypes.MatchString(groupVersionKind.Kind) {
					if err := AddTknResource(groupVersionKind.Kind, []byte(f), &tknRes); err != nil {
						return tknRes, errors.Wrapf(err, "failed to parse resources")
					}
				}
			}
		} else {
			return tknRes, errors.Wrapf(err, "error parsing yaml file")
		}
	} else {
		return tknRes, errors.New("invalid input file")
	}
	return tknRes, nil
}

func AddTknResource(kind string, objDataBuf []byte, raw *common.TknResources) error {
	switch kind {
	case "Task":
		task, _ := parseTknTask(objDataBuf)
		//Handle error
		raw.TaskSpecs = append(raw.TaskSpecs, task)
	case "Pipeline":
		pipeline, _ := parseTknPipeline(objDataBuf)
		//Handle error
		raw.PipelineSpecs = append(raw.PipelineSpecs, pipeline)
	case "TaskRun":
		taskrun, _ := parseTknTaskRun(objDataBuf)
		//Handle error
		raw.TaskrunSpecs = append(raw.TaskrunSpecs, taskrun)
	case "PipelineRun":
		pipelinerun, _ := parseTknPipelineRun(objDataBuf)
		//Handle error
		raw.PipelinerunSpecs = append(raw.PipelinerunSpecs, pipelinerun)
	}

	return nil
}

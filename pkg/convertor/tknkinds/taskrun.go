package tknkinds

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	toto "github.com/in-toto/in-toto-golang/in_toto"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	v1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

const (
	commitParam        = "CHAINS-GIT_COMMIT"
	urlParam           = "CHAINS-GIT_URL"
	ociDigestResult    = "IMAGE_DIGEST"
	chainsDigestSuffix = "_DIGEST"
)

//TaskrunToIntotoProvenance :
func TaskrunToIntotoProvenance(taskrun *v1beta1.TaskRun) ([]byte, error) {
	currenttime := time.Now()
	currenttime = currenttime.Add(30 * 24 * time.Hour)

	var keys = make(map[string]toto.Key)

	intotoSteps, err := getIntotoSteps(taskrun)
	if err != nil {
		return nil, fmt.Errorf("failed to get steps in taskrun %s: %v", taskrun.Name, err)
	}
	intotoInspect, err := getIntotoInspect(taskrun)
	if err != nil {
		return nil, fmt.Errorf("failed to get inspections steps in pipelinerun %s: %v", taskrun.Name, err)
	}
	var metablock = toto.Metablock{
		Signed: toto.Layout{
			Type:    taskrun.Name,
			Expires: currenttime.Format("2006-01-02T15:04:05Z"),
			Steps:   intotoSteps,
			Inspect: intotoInspect,
			Keys:    keys,
		},
	}

	return json.MarshalIndent(metablock, "", "    ")
}

func container(stepState v1beta1.StepState, tr *v1beta1.TaskRun) v1beta1.Step {
	name := stepState.Name
	if tr.Status.TaskSpec != nil {
		for _, s := range tr.Status.TaskSpec.Steps {
			if s.Name == name {
				return s
			}
		}
	}

	return v1beta1.Step{}
}

// Get the taskruns from a pipelinerun append together to be passed into intoto layout
func getIntotoSteps(taskrun *v1beta1.TaskRun) ([]toto.Step, error) {
	var intotoSteps []toto.Step
	var expMats [][]string

	gitCommit, gitURL := gitInfo(taskrun)

	// Store git rev as Materials and Recipe.Material
	if gitCommit != "" && gitURL != "" {
		expMats = append(expMats, [][]string{{"ALLOW", gitURL}}...)
	}

	params := getParams(taskrun)

	for _, step := range taskrun.Status.Steps {

		intotoStep := toto.Step{
			Type:            "step",
			ExpectedCommand: GetExpectedCommand(container(step, taskrun), params),
			SupplyChainItem: toto.SupplyChainItem{
				Name:              step.Name,
				ExpectedMaterials: append(expMats, GetExpectedMaterials(step)...),
				ExpectedProducts:  getExpectedProducts(taskrun, false),
			},
		}
		intotoSteps = append(intotoSteps, intotoStep)
	}
	return intotoSteps, nil
}

// Get the taskruns from a pipelinerun append together for the inspection section of the layout
func getIntotoInspect(taskrun *v1beta1.TaskRun) ([]toto.Inspection, error) {
	var intotoInsepcts []toto.Inspection

	for _, step := range taskrun.Status.Steps {

		intotoInspect := toto.Inspection{
			Type: "Inspect",
			Run:  []string{""},
			SupplyChainItem: toto.SupplyChainItem{
				Name:              step.Name,
				ExpectedMaterials: getExpectedProducts(taskrun, true),
				//ExpectedProducts:  getExpectedProducts(task),
			},
		}
		intotoInsepcts = append(intotoInsepcts, intotoInspect)
	}

	return intotoInsepcts, nil
}

// Get the commands and arguments from a pipeline step and append together to be passed into intoto layout
func GetExpectedCommand(step v1beta1.Step, params map[string]string) []string {
	var combined []string

	// combine the command and arguments into one
	combined = append(combined, step.Command...)
	fmt.Println("params", params)
	for _, arg := range step.Args {
		if strings.Contains(arg, "params") {
			if strings.Contains(arg, "=") {
				argSplit := strings.Split(arg, "=")
				t := strings.ReplaceAll(argSplit[1], "$(params.", "")
				t = strings.ReplaceAll(t, ")", "")
				fmt.Println("arg Value: ", t)
				if params[t] != "" {
					argComplete := argSplit[0] + "=" + params[t]
					combined = append(combined, argComplete)
				}
			} else {
				t := strings.ReplaceAll(arg, "$(params.", "")
				t = strings.ReplaceAll(t, ")", "")
				fmt.Println("arg Value: ", t)
				if params[t] != "" {
					argComplete := params[t]
					combined = append(combined, argComplete)
				}
			}

		} else {
			combined = append(combined, arg)
		}
	}

	return combined
}

func gitInfo(tr *v1beta1.TaskRun) (commit string, url string) {
	// Scan for git params to use for materials
	for _, p := range tr.Spec.Params {
		if p.Name == commitParam {
			commit = p.Value.StringVal
			continue
		}
		if p.Name == urlParam {
			url = p.Value.StringVal
			// make sure url is PURL (git+https)
			if !strings.HasPrefix(url, "git+") {
				url = "git+" + url
			}
		}
	}
	return
}

// Get the materials from a pipeline task and append together to be passed into intoto layout
func GetExpectedMaterials(trs v1beta1.StepState) [][]string {
	var expMats [][]string

	// Get each input to be placed into intoto layout file

	uri := getPackageURLDocker(trs.ImageID)
	if uri != "" {
		expMats = append(expMats, [][]string{{"ALLOW", uri}}...)
	}

	// Append disallow statement to the end of the expected materials
	expMats = append(expMats, [][]string{{"DISALLOW", "*"}}...)

	return expMats
}

// Get parameters from tasks that can be substituted into the commands and arguments
func getParams(tr *v1beta1.TaskRun) map[string]string {

	// get parameters
	params := make(map[string]string)
	for _, p := range tr.Spec.Params {
		params[p.Name] = p.Value.StringVal
	}

	return params
}

// getPackageURLDocker takes an image id and creates a package URL string
// based from it.
// https://github.com/package-url/purl-spec
func getPackageURLDocker(imageID string) string {
	// Default registry per purl spec
	const defaultRegistry = "hub.docker.com"

	// imageID formats: name@alg:digest
	//                  schema://name@alg:digest
	d := strings.Split(imageID, "//")
	if len(d) == 2 {
		// Get away with schema
		imageID = d[1]
	}

	digest, err := name.NewDigest(imageID, name.WithDefaultRegistry(defaultRegistry))
	if err != nil {
		return ""
	}

	// DigestStr() is alg:digest
	parts := strings.Split(digest.DigestStr(), ":")
	if len(parts) != 2 {
		return ""
	}

	purl := fmt.Sprintf("pkg:docker/%s@%s",
		digest.Context().RepositoryStr(),
		digest.DigestStr(),
	)

	// Only inlude registry if it's not the default
	registry := digest.Context().Registry.Name()
	if registry != defaultRegistry {
		purl = fmt.Sprintf("%s?repository_url=%s", purl, registry)
	}

	return purl
}

// getOCIImageID generates an imageID that is compatible imageID field in
// the task result's status field.
func getOCIImageID(name, alg, digest string) string {
	// image id is: docker://name@alg:digest
	return fmt.Sprintf("docker://%s@%s:%s", name, alg, digest)
}

// Get the prodcuts from a pipeline taskrun and append together to be passed into intoto layout
func getExpectedProducts(tr *v1beta1.TaskRun, inspection bool) [][]string {
	var expProds [][]string

	// If not part of the inspection step within the intoto layout grab the matierlas from the input
	if !inspection {

		for _, trr := range tr.Status.TaskRunResults {
			if !strings.HasSuffix(trr.Name, chainsDigestSuffix) {
				continue
			}
			potentialKey := strings.TrimSuffix(trr.Name, chainsDigestSuffix)
			var sub string
			// try to match the key to a param or a result
			for _, p := range tr.Spec.Params {
				if potentialKey != p.Name || p.Value.Type != v1beta1.ParamTypeString {
					continue
				}
				sub = p.Value.StringVal
			}
			for _, p := range tr.Status.TaskRunResults {
				if potentialKey == p.Name {
					sub = strings.TrimRight(p.Value, "\n")
				}
			}
			// if we couldn't match to a param or a result, continue
			if sub == "" {
				continue
			}
			// Value should be of the format:
			//   alg:hash
			//   alg:hash filename
			algHash := strings.Split(trr.Value, " ")[0]
			ah := strings.Split(algHash, ":")
			if len(ah) != 2 {
				continue
			}
			alg := ah[0]
			hash := ah[1]

			// OCI image shall use pacakge url format for subjects
			if trr.Name == ociDigestResult {
				imageID := getOCIImageID(sub, alg, hash)
				sub = getPackageURLDocker(imageID)
			}
			expProds = append(expProds, [][]string{{"ALLOW", sub}}...)

		}
		// Append disallow statement to the end of the expected prodcuts
		//expProds = append(expProds, [][]string{{"DISALLOW", "*"}}...)
		return expProds

	} else {
		// Get each output to be placed into intoto layout file for inspection

		if tr.Spec.Resources == nil {
			return expProds
		}

		// go through resourcesResult
		if tr.Spec.Resources != nil {
			for _, output := range tr.Spec.Resources.Outputs {
				name := output.Name
				if output.PipelineResourceBinding.ResourceSpec == nil {
					continue
				}
				if output.PipelineResourceBinding.ResourceSpec.Type == v1alpha1.PipelineResourceTypeImage {
					for _, s := range tr.Status.ResourcesResult {
						if s.ResourceName == name {
							if s.Key == "url" {
								expProds = append(expProds, [][]string{{"MATCH", s.Value,
									"WITH", "PRODUCTS", "FROM", tr.Name}}...)

							}
						}
					}
				}
			}
		}
		return expProds
	}

}

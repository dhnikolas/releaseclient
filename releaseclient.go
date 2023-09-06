package releaseclient

import (
	"context"
	"encoding/base64"
	releasev1alpha1 "github.com/dhnikolas/release-operator/api/v1alpha1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReleaseClient struct {
	kubeClient     client.Client
	ignoreServices []string
}

func New(kubeConfigBase64 string, ignoreServices []string) (*ReleaseClient, error) {
	kubeConfig, err := base64.StdEncoding.DecodeString(kubeConfigBase64)
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	c, err := client.New(config, client.Options{})
	if err != nil {
		return nil, err
	}

	return &ReleaseClient{
		kubeClient:     c,
		ignoreServices: ignoreServices,
	}, nil
}

type Task struct {
	BranchName string
	Services   []string
}

func (r *ReleaseClient) ApplyBuild(ctx context.Context, buildName string, taskList []Task) {
	build := &releasev1alpha1.Build{}
	build.Namespace = "default"
	build.Name = buildName

	repos := new([]releasev1alpha1.Repo)
	for _, task := range taskList {
		for _, service := range task.Services {

		}
	}

	build.Spec.Repos

}

func addBranchToRepo(repos *[]releasev1alpha1.Repo, repoName, branchName string) {
	currentRepo := getRepoByName(repos, repoName)
	if currentRepo == nil {
		*repos = append(*repos, releasev1alpha1.Repo{
			URL:      repoName,
			Branches: []releasev1alpha1.Branch{{Name: branchName}},
		})
	} else {
		currentRepo.Branches = append(currentRepo.Branches)
	}
}

func getRepoByName(repos *[]releasev1alpha1.Repo, repoName string) *releasev1alpha1.Repo {
	for _, repo := range *repos {
		if repo.URL == repoName {
			return &repo
		}
	}
	return nil
}

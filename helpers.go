package releaseclient

import (
	releasev1alpha1 "github.com/dhnikolas/release-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	crc "sigs.k8s.io/controller-runtime/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"context"
)

func taskToRepo(taskList []Task) []releasev1alpha1.Repo {
	repos := new([]releasev1alpha1.Repo)
	for _, task := range taskList {
		for _, service := range task.Services {
			addBranchToRepo(repos, service, task.BranchName)
		}
	}

	return *repos
}

func addBranchToRepo(repos *[]releasev1alpha1.Repo, repoName, branchName string) {
	repoIndex, ok := getRepoByName(repos, repoName)
	if !ok {
		*repos = append(*repos, releasev1alpha1.Repo{
			URL:      repoName,
			Branches: []releasev1alpha1.Branch{{Name: branchName}},
		})
	} else {
		currentRepo := *repos
		currentRepo[repoIndex].Branches = append(currentRepo[repoIndex].Branches, releasev1alpha1.Branch{Name: branchName})
	}
}

func getRepoByName(repos *[]releasev1alpha1.Repo, repoName string) (int, bool) {
	for i, repo := range *repos {
		if repo.URL == repoName {
			return i, true
		}
	}
	return 0, false
}

func BuildScheme(builders ...*scheme.Builder) *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	for _, builder := range builders {
		utilruntime.Must(builder.AddToScheme(s))
	}
	return s
}

func getObject(ctx context.Context, c crc.Client, name string, obj client.Object) (bool, error) {
	err := c.Get(ctx, client.ObjectKey{Name: name, Namespace: obj.GetNamespace()}, obj)
	if err != nil {
		switch v := err.(type) {
		case apierrors.APIStatus:
			if v.Status().Code == 404 {
				return false, nil
			}
			return false, err
		default:
			return false, err
		}
	}
	return true, nil
}

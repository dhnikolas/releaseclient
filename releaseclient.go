package releaseclient

import (
	"context"
	"encoding/base64"
	releasev1alpha1 "github.com/dhnikolas/release-operator/api/v1alpha1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crc "sigs.k8s.io/controller-runtime/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReleaseClient struct {
	kubeClient     client.Client
	ignoreServices []string
	namespace      string
}

type Task struct {
	BranchName string
	Services   []string
}

type BuildStatus struct {
	Name  string
	Ready bool
	Tasks []*TaskStatus
}

type TaskStatus struct {
	BranchName string
	Statuses   []BranchStatus
}

type BranchStatus struct {
	ServiceName    string
	Valid          bool
	Merged         bool
	Conflict       bool
	ConflictBranch string
}

func New(kubeConfigBase64, defaultNamespace string, ignoreServices []string) (*ReleaseClient, error) {
	kubeConfig, err := base64.StdEncoding.DecodeString(kubeConfigBase64)
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeConfig)
	c, err := client.New(config, client.Options{
		Scheme: BuildScheme(releasev1alpha1.SchemeBuilder),
	})
	if err != nil {
		return nil, err
	}

	return &ReleaseClient{
		kubeClient:     c,
		ignoreServices: ignoreServices,
		namespace:      defaultNamespace,
	}, nil
}

func (bs *BuildStatus) GetTaskByName(name string) *TaskStatus {
	for _, task := range bs.Tasks {
		if task.BranchName == name {
			return task
		}
	}
	return nil
}

func (r *ReleaseClient) ApplyBuild(ctx context.Context, buildName string, taskList []Task) error {
	build := &releasev1alpha1.Build{}
	build.Namespace = r.namespace
	build.Name = buildName
	exist, err := getObject(ctx, r.kubeClient, buildName, build)
	if err != nil {
		return err
	}
	repos := taskToRepo(taskList)
	build.Spec.Repos = repos
	if exist {
		err = r.kubeClient.Update(ctx, build)
		if err != nil {
			return err
		}

	} else {
		err := r.kubeClient.Create(ctx, build)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReleaseClient) GetBuildInfo(ctx context.Context, buildName string) (*BuildStatus, error) {
	build := &releasev1alpha1.Build{}
	build.Namespace = r.namespace
	build.Name = buildName
	_, err := getObject(ctx, r.kubeClient, buildName, build)
	if err != nil {
		return nil, err
	}

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "build.release.salt.x5.ru/build-name",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{buildName},
			},
		},
	})

	merges := &releasev1alpha1.MergeList{}
	err = r.kubeClient.List(ctx, merges, &crc.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}

	bs := &BuildStatus{
		Name:  buildName,
		Tasks: make([]*TaskStatus, 0),
	}
	isReady := true
	for _, item := range merges.Items {
		for _, branch := range item.Status.Branches {
			if branch.IsMerged != "True" {
				isReady = false
			}
			taskStatus := bs.GetTaskByName(branch.Name)
			resolveBranch := ""
			if item.Status.ResolveConflictBranch != nil {
				resolveBranch = item.Status.ResolveConflictBranch.Name
			}
			if taskStatus == nil {
				bs.Tasks = append(bs.Tasks, &TaskStatus{
					BranchName: branch.Name,
					Statuses: []BranchStatus{
						{
							ServiceName:    item.Status.ProjectPID,
							Valid:          branch.IsValid == "True",
							Merged:         branch.IsMerged == "True",
							Conflict:       item.Status.ResolveConflictBranch != nil,
							ConflictBranch: resolveBranch,
						},
					},
				})
			} else {
				taskStatus.Statuses = append(taskStatus.Statuses, BranchStatus{
					ServiceName:    item.Status.ProjectPID,
					Valid:          branch.IsValid == "True",
					Merged:         branch.IsMerged == "True",
					Conflict:       item.Status.ResolveConflictBranch != nil,
					ConflictBranch: resolveBranch,
				})
			}
		}
	}

	bs.Ready = isReady
	return bs, nil
}

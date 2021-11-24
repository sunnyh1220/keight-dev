package sample

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

type Args struct {
	FavoriteColor  string `json:"favoriteColor"`
	FavoriteNumber int    `json:"favoriteNumber"`
	ThanksTo       string `json:"thanksTo"`
}

type Sample struct {
	args   *Args
	handle framework.FrameworkHandle
}

const (
	Name              = "sample-plugin"
	preFilterStateKey = "PreFilter" + Name
)

// 实现PreFilter，Filter扩展点
var _ framework.PreFilterPlugin = &Sample{}
var _ framework.FilterPlugin = &Sample{}

func (s *Sample) Filter(ctx context.Context, state *framework.CycleState, pod *corev1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	preState, err := getPreFilterState(state)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}
	if klog.V(2).Enabled() {
		klog.InfoS("Start Filter Pod", "pod", pod.Name, "node", nodeInfo.Node().Name, "preFilterState", preState)
	}
	// logic

	return framework.NewStatus(framework.Success, "")
}

func getPreFilterState(state *framework.CycleState) (*preFilterState, error) {
	data, err := state.Read(preFilterStateKey)
	if err != nil {
		return nil, err
	}
	s, ok := data.(*preFilterState)
	if !ok {
		return nil, fmt.Errorf("%+v convert to SamplePlugin preFilterState error", data)
	}

	return s, nil
}

func (s *Sample) PreFilter(ctx context.Context, state *framework.CycleState, pod *corev1.Pod) *framework.Status {
	if klog.V(2).Enabled() {
		klog.InfoS("Start PreFilter Pod", "pod", pod.Name)
	}
	state.Write(preFilterStateKey, computerPodResourceLimit(pod))
	return nil
}

type preFilterState struct {
	framework.Resource // requests,limits
}

func (p *preFilterState) Clone() framework.StateData {
	return p
}

func computerPodResourceLimit(pod *corev1.Pod) *preFilterState {
	result := &preFilterState{}
	for _, container := range pod.Spec.Containers {
		result.Add(container.Resources.Limits)
	}
	return result
}

func (s *Sample) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

func (s *Sample) Name() string {
	return Name
}

func New(object runtime.Object, f framework.FrameworkHandle) (framework.Plugin, error) {
	args, err := getSampleArgs(object)
	if err != nil {
		return nil, err
	}
	// validate args
	if klog.V(2).Enabled() {
		klog.InfoS("Successfully get plugin config args", "plugin", Name, "args", args)
	}

	return &Sample{
		args:   args,
		handle: f,
	}, nil
}

func getSampleArgs(object runtime.Object) (*Args, error) {
	sa := &Args{}
	if err := frameworkruntime.DecodeInto(object, sa); err != nil {
		return nil, err
	}
	return sa, nil
}

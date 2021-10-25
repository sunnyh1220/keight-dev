package controllers

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Action 定义的执行动作接口
type Action interface {
	Execute(context.Context) error
}

// PatchStatus 更新资源对象status
type PatchStatus struct {
	client   client.Client
	original runtime.Object
	new      client.Object
}

func (s *PatchStatus) Execute(ctx context.Context) error {
	if reflect.DeepEqual(s.original, s.new) {
		return nil
	}

	if err := s.client.Status().Patch(ctx, s.new, client.MergeFrom(s.original)); err != nil {
		return fmt.Errorf("while patching status error %q", err)
	}

	return nil
}

// CreateObject 创建一个新的资源对象
type CreateObject struct {
	client client.Client
	obj    client.Object
}

func (o *CreateObject) Execute(ctx context.Context) error {
	if err := o.client.Create(ctx, o.obj); err != nil {
		return fmt.Errorf("error %q while creating object ", err)
	}
	return nil
}

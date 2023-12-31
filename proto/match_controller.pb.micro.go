// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: proto/match_controller.proto

package match_controller

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

import (
	context "context"
	api "github.com/micro/micro/v3/service/api"
	client "github.com/micro/micro/v3/service/client"
	server "github.com/micro/micro/v3/service/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ api.Endpoint
var _ context.Context
var _ client.Option
var _ server.Option

// Api Endpoints for MatchController service

func NewMatchControllerEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for MatchController service

type MatchControllerService interface {
	AddTask(ctx context.Context, in *AddTaskReq, opts ...client.CallOption) (*AddTaskRsp, error)
}

type matchControllerService struct {
	c    client.Client
	name string
}

func NewMatchControllerService(name string, c client.Client) MatchControllerService {
	return &matchControllerService{
		c:    c,
		name: name,
	}
}

func (c *matchControllerService) AddTask(ctx context.Context, in *AddTaskReq, opts ...client.CallOption) (*AddTaskRsp, error) {
	req := c.c.NewRequest(c.name, "MatchController.AddTask", in)
	out := new(AddTaskRsp)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for MatchController service

type MatchControllerHandler interface {
	AddTask(context.Context, *AddTaskReq, *AddTaskRsp) error
}

func RegisterMatchControllerHandler(s server.Server, hdlr MatchControllerHandler, opts ...server.HandlerOption) error {
	type matchController interface {
		AddTask(ctx context.Context, in *AddTaskReq, out *AddTaskRsp) error
	}
	type MatchController struct {
		matchController
	}
	h := &matchControllerHandler{hdlr}
	return s.Handle(s.NewHandler(&MatchController{h}, opts...))
}

type matchControllerHandler struct {
	MatchControllerHandler
}

func (h *matchControllerHandler) AddTask(ctx context.Context, in *AddTaskReq, out *AddTaskRsp) error {
	return h.MatchControllerHandler.AddTask(ctx, in, out)
}

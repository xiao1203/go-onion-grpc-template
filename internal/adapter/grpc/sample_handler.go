package grpc

import (
	"context"

	"connectrpc.com/connect"
	samplev1 "github.com/xiao1203/go-onion-grpc-template/gen/sample/v1"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain"
	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"
)

type SampleHandler struct {
	uc *usecase.SampleUsecase
}

func NewSampleHandler(uc *usecase.SampleUsecase) *SampleHandler {
	return &SampleHandler{uc: uc}
}

func (h *SampleHandler) CreateSample(
	ctx context.Context,
	req *connect.Request[samplev1.CreateSampleRequest],
) (*connect.Response[samplev1.CreateSampleResponse], error) {
	in := &domain.Sample{
		Name:    req.Msg.GetName(),
		Content: req.Msg.GetContent(),
		Count:   req.Msg.GetCount(),
	}
	out, err := h.uc.Create(ctx, in)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := connect.NewResponse(&samplev1.CreateSampleResponse{
		Sample: toProtoSample(out),
	})
	return res, nil
}

func (h *SampleHandler) GetSample(
	ctx context.Context,
	req *connect.Request[samplev1.GetSampleRequest],
) (*connect.Response[samplev1.GetSampleResponse], error) {
	out, err := h.uc.Get(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&samplev1.GetSampleResponse{Sample: toProtoSample(out)}), nil
}

func (h *SampleHandler) ListSamples(
	ctx context.Context,
	req *connect.Request[samplev1.ListSamplesRequest],
) (*connect.Response[samplev1.ListSamplesResponse], error) {
	// NOTE: proto currently has no paging fields. Pass defaults.
	items, err := h.uc.List(ctx, domain.ListParams{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := make([]*samplev1.Sample, 0, len(items))
	for _, it := range items {
		out = append(out, toProtoSample(it))
	}
	return connect.NewResponse(&samplev1.ListSamplesResponse{Samples: out}), nil
}

func (h *SampleHandler) UpdateSample(
	ctx context.Context,
	req *connect.Request[samplev1.UpdateSampleRequest],
) (*connect.Response[samplev1.UpdateSampleResponse], error) {
	in := &domain.Sample{
		ID:      req.Msg.GetId(),
		Name:    req.Msg.GetName(),
		Content: req.Msg.GetContent(),
		Count:   req.Msg.GetCount(),
	}
	out, err := h.uc.Update(ctx, in)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&samplev1.UpdateSampleResponse{Sample: toProtoSample(out)}), nil
}

func (h *SampleHandler) DeleteSample(
	ctx context.Context,
	req *connect.Request[samplev1.DeleteSampleRequest],
) (*connect.Response[samplev1.DeleteSampleResponse], error) {
	if err := h.uc.Delete(ctx, req.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&samplev1.DeleteSampleResponse{}), nil
}

func toProtoSample(in *domain.Sample) *samplev1.Sample {
	if in == nil {
		return nil
	}
	return &samplev1.Sample{
		Id:      in.ID,
		Name:    in.Name,
		Content: in.Content,
		Count:   in.Count,
	}
}

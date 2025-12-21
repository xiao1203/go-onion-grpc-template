package grpc

import (
	"context"

	"connectrpc.com/connect"
	userv1 "github.com/xiao1203/go-onion-grpc-template/gen/user/v1"
	"github.com/xiao1203/go-onion-grpc-template/internal/auth"
	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"
)

type UserHandler struct{ uc *usecase.UserUsecase }

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler { return &UserHandler{uc: uc} }

func (h *UserHandler) GetMe(ctx context.Context, req *connect.Request[userv1.GetMeRequest]) (*connect.Response[userv1.GetMeResponse], error) {
	p, ok := auth.FromContext(ctx)
	if !ok || p.UserID == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}
	u, err := h.uc.GetMe(ctx, p.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := &userv1.GetMeResponse{User: toProtoUser(u)}
	return connect.NewResponse(res), nil
}

func (h *UserHandler) UpdateMyProfile(ctx context.Context, req *connect.Request[userv1.UpdateMyProfileRequest]) (*connect.Response[userv1.UpdateMyProfileResponse], error) {
	p, ok := auth.FromContext(ctx)
	if !ok || p.UserID == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}
	u, err := h.uc.UpdateMyProfile(ctx, p.UserID, req.Msg.GetDisplayName(), req.Msg.GetPictureUrl())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := &userv1.UpdateMyProfileResponse{User: toProtoUser(u)}
	return connect.NewResponse(res), nil
}

func toProtoUser(u *usecase.User) *userv1.User {
	if u == nil {
		return nil
	}
	return &userv1.User{
		Id:          uint64(u.ID),
		Email:       u.Email,
		DisplayName: u.DisplayName,
		PictureUrl:  u.PictureURL,
		Roles:       append([]string(nil), u.Roles...),
	}
}

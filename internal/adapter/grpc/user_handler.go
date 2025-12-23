package grpc

import (
    "context"

    "connectrpc.com/connect"
    "github.com/newmo-oss/ergo"
    userv1 "github.com/xiao1203/go-onion-grpc-template/gen/user/v1"
    "github.com/xiao1203/go-onion-grpc-template/internal/apperr"
    "github.com/xiao1203/go-onion-grpc-template/internal/auth"
    "github.com/xiao1203/go-onion-grpc-template/internal/domain/entity"
    "github.com/xiao1203/go-onion-grpc-template/internal/usecase"
)

type UserHandler struct{ uc *usecase.UserUsecase }

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler { return &UserHandler{uc: uc} }

func (h *UserHandler) GetMe(ctx context.Context, req *connect.Request[userv1.GetMeRequest]) (*connect.Response[userv1.GetMeResponse], error) {
    p, ok := auth.FromContext(ctx)
    if !ok || p.UserID == 0 {
        return nil, apperr.ToConnect(ergo.WithCode(ergo.New("unauthenticated"), apperr.Unauthenticated))
    }
    u, err := h.uc.GetMe(ctx, p.UserID)
    if err != nil {
        return nil, apperr.ToConnect(err)
    }
	res := &userv1.GetMeResponse{User: toProtoUser(u)}
	return connect.NewResponse(res), nil
}

func (h *UserHandler) UpdateMyProfile(ctx context.Context, req *connect.Request[userv1.UpdateMyProfileRequest]) (*connect.Response[userv1.UpdateMyProfileResponse], error) {
    p, ok := auth.FromContext(ctx)
    if !ok || p.UserID == 0 {
        return nil, apperr.ToConnect(ergo.WithCode(ergo.New("unauthenticated"), apperr.Unauthenticated))
    }
    u, err := h.uc.UpdateMyProfile(ctx, p.UserID, req.Msg.GetDisplayName(), req.Msg.GetPictureUrl())
    if err != nil {
        return nil, apperr.ToConnect(err)
    }
	res := &userv1.UpdateMyProfileResponse{User: toProtoUser(u)}
	return connect.NewResponse(res), nil
}

func toProtoUser(u *entity.User) *userv1.User {
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

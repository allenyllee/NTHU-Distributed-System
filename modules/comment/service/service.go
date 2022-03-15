package service

import (
	"context"
	"errors"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/pb"
	videopb "github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/google/uuid"
)

type service struct {
	pb.UnimplementedCommentServer

	commentDAO  dao.CommentDAO
	videoClient videopb.VideoClient
}

func NewService(commentDAO dao.CommentDAO, videoClient videopb.VideoClient) *service {
	return &service{
		commentDAO:  commentDAO,
		videoClient: videoClient,
	}
}

func (s *service) Healthz(ctx context.Context, req *pb.HealthzRequest) (*pb.HealthzResponse, error) {
	return &pb.HealthzResponse{Status: "ok"}, nil
}

/*
gRPC TODO:
1. Call dao API to list comment and check if any error happen. If so, return an nil response.
2. Create an array to store the return value. Remember to transform the data schema from dao into protobuf.
   (Refer to the ToProto method defined in dao/comment.go. This method is used to transform data schema from dao into protobuf.)
3. Pack the array into correct format.
*/
func (s *service) ListComment(ctx context.Context, req *pb.ListCommentRequest) (*pb.ListCommentResponse, error) {
	comments, err := s.commentDAO.ListByVideoID(ctx, req.GetVideoId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}

	pbComments := make([]*pb.CommentInfo, 0, len(comments))
	for _, comment := range comments {
		pbComments = append(pbComments, comment.ToProto())
	}

	return &pb.ListCommentResponse{Comments: pbComments}, nil
}

/*
gRPC TODO:
1. Send gRPC to video server to check if the video id in request is valid. Think about which gRPC provided by video server to use?
   If any error happened, return nil response and the error.
2. Create a comment with information in request.
3. Call dao API to create a new comment and do error handling.
4. Return the result. You may use .String() method to transform the return value of dao API to a string.
*/
func (s *service) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CreateCommentResponse, error) {
	_, err := s.videoClient.GetVideo(ctx, &videopb.GetVideoRequest{Id: req.GetVideoId()})
	if err != nil {
		return nil, err
	}

	commentID, err := s.commentDAO.Create(ctx, &dao.Comment{VideoID: req.GetVideoId(), Content: req.GetContent()})
	if err != nil {
		return nil, err
	}

	return &pb.CreateCommentResponse{Id: commentID.String()}, nil
}

/*
gRPC TODO:
1. Update a comment with information in request.
2. Call dao API to update a comment and do error handling. You need to handle comment not found error and other unknown error here.
3. Return the result. Don't forget to transform the data schema from dao into proto.
   (Refer to the ToProto method defined in dao/comment.go. This method is used to transform data schema from dao into protobuf.)
*/
func (s *service) UpdateComment(ctx context.Context, req *pb.UpdateCommentRequest) (*pb.UpdateCommentResponse, error) {
	commentID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, ErrInvalidUUID
	}

	comment := dao.Comment{ID: commentID, Content: req.GetContent()}

	if err := s.commentDAO.Update(ctx, &comment); err != nil {
		if errors.Is(err, dao.ErrCommentNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	return &pb.UpdateCommentResponse{Comment: comment.ToProto()}, nil
}

/*
gRPC TODO:
1. Call dao API to delete a comment and do error handling. You need to handle comment not found error and other unknown error here.
2. Return the response.
*/
func (s *service) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteCommentResponse, error) {
	commentID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, ErrInvalidUUID
	}

	if err := s.commentDAO.Delete(ctx, commentID); err != nil {
		if errors.Is(err, dao.ErrCommentNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	return &pb.DeleteCommentResponse{}, nil
}

/*
gRPC TODO:
1. Call dao API to delete comments by video id and do error handling. You need to do error handling here.
2. Return the response.
*/
func (s *service) DeleteCommentByVideoID(ctx context.Context, req *pb.DeleteCommentByVideoIDRequest) (*pb.DeleteCommentByVideoIDResponse, error) {
	if err := s.commentDAO.DeleteByVideoID(ctx, req.GetVideoId()); err != nil {
		if errors.Is(err, dao.ErrCommentNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	return &pb.DeleteCommentByVideoIDResponse{}, nil
}

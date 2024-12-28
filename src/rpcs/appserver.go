package rpcs

import (
	"context"

	"mist/src/middleware"
	pb_mistbe "mist/src/protos/mistbe/v1"
	"mist/src/service"
)

func (s *Grpcserver) CreateAppserver(
	ctx context.Context, req *pb_mistbe.CreateAppserverRequest,
) (*pb_mistbe.CreateAppserverResponse, error) {
	appserverService := service.NewAppserverService(s.DbcPool, ctx)
	jwtClaims, _ := middleware.GetJWTClaims(ctx)

	appserver, err := appserverService.Create(req.GetName(), jwtClaims.UserID)

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_mistbe.CreateAppserverResponse{
		Appserver: appserverService.PgTypeToPb(appserver),
	}, nil
}

func (s *Grpcserver) GetByIdAppserver(
	ctx context.Context, req *pb_mistbe.GetByIdAppserverRequest,
) (*pb_mistbe.GetByIdAppserverResponse, error) {
	appserverService := service.NewAppserverService(s.DbcPool, ctx)
	appserver, err := appserverService.GetById(req.GetId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_mistbe.GetByIdAppserverResponse{Appserver: appserverService.PgTypeToPb(appserver)}, nil
}

func (s *Grpcserver) ListAppservers(
	ctx context.Context, req *pb_mistbe.ListAppserversRequest,
) (*pb_mistbe.ListAppserversResponse, error) {
	appserverService := service.NewAppserverService(s.DbcPool, ctx)
	// TODO: Figure out what can go wrong to add error handler
	appservers, _ := appserverService.List(req.GetName())

	response := &pb_mistbe.ListAppserversResponse{}

	// Resize the array
	response.Appservers = make([]*pb_mistbe.Appserver, 0, len(appservers))

	for _, appserver := range appservers {
		response.Appservers = append(response.Appservers, appserverService.PgTypeToPb(&appserver))
	}

	return response, nil
}

func (s *Grpcserver) DeleteAppserver(
	ctx context.Context, req *pb_mistbe.DeleteAppserverRequest,
) (*pb_mistbe.DeleteAppserverResponse, error) {

	err := service.NewAppserverService(s.DbcPool, ctx).Delete(req.GetId())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_mistbe.DeleteAppserverResponse{}, nil
}

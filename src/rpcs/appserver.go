package rpcs

import (
	"context"
	"log"

	pb_mistbe "mist/src/protos/mistbe/v1"
	"mist/src/service"
)

func (s *GRPCServer) CreateAppserver(
	ctx context.Context, req *pb_mistbe.CreateAppserverRequest,
) (*pb_mistbe.CreateAppserverResponse, error) {
	appserver_service := service.NewAppserverService(s.dbc_pool, ctx)
	appserver, err := appserver_service.Create(req.GetName())

	if err != nil {
		return nil, ErrorHandler(err)
	}

	return &pb_mistbe.CreateAppserverResponse{
		Appserver: appserver_service.PgTypeToPb(appserver),
	}, nil
}

func (s *GRPCServer) GetByIdAppserver(
	ctx context.Context, req *pb_mistbe.GetByIdAppserverRequest,
) (*pb_mistbe.GetByIdAppserverResponse, error) {
	appserver_service := service.NewAppserverService(s.dbc_pool, ctx)
	appserver, err := appserver_service.GetById(req.GetId())

	if err != nil {
		// TODO: handle this error
	}

	return &pb_mistbe.GetByIdAppserverResponse{Appserver: appserver_service.PgTypeToPb(&appserver)}, nil
}

func (s *GRPCServer) ListAppservers(
	ctx context.Context, req *pb_mistbe.ListAppserversRequest,
) (*pb_mistbe.ListAppserversResponse, error) {
	appserver_service := service.NewAppserverService(s.dbc_pool, ctx)
	appservers, err := appserver_service.List(req.GetName())

	if err != nil {
		// TODO: handle this error
	}
	response := &pb_mistbe.ListAppserversResponse{}

	// Resize the array
	response.Appservers = make([]*pb_mistbe.Appserver, 0, len(appservers))

	for _, appserver := range appservers {
		response.Appservers = append(response.Appservers, appserver_service.PgTypeToPb(&appserver))
	}

	return response, nil
}

func (s *GRPCServer) DeleteAppserver(
	ctx context.Context, req *pb_mistbe.DeleteAppserverRequest,
) (*pb_mistbe.DeleteAppserverResponse, error) {

	err := service.NewAppserverService(s.dbc_pool, ctx).Delete(req.GetId())

	if err != nil {
		log.Printf("Failure deleting: %v", err.Error())
		// TODO: handle this error
	}

	return &pb_mistbe.DeleteAppserverResponse{}, nil
}

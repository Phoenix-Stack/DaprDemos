package service

import (
	"daprdemos/golang/customer/config/db"
	"daprdemos/golang/customer/models"
	pb "daprdemos/golang/customer/protos/customer_v1"
)

type CustomerService struct {
}

func (s *CustomerService) GetCustomerById(req *pb.IdRequest) pb.Customer {
	var customer models.Customer
	db.DB.First(&customer, "id = ?", req.Id)
	return pb.Customer{
		Id:   customer.ID.String(),
		Name: customer.Name,
	}
}

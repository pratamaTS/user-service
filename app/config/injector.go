//go:build wireinject
// +build wireinject

package config

import (
	"github.com/google/wire"

	"harjonan.id/user-service/app/controller"
	r2 "harjonan.id/user-service/app/infra/r2"
	"harjonan.id/user-service/app/repository"
	"harjonan.id/user-service/app/service"
)

var db = wire.NewSet(ConnectToMongoDB)

var r2Set = wire.NewSet(
	r2.MustNew,
)

var repoSet = wire.NewSet(
	repository.AdminRepositoryInit, wire.Bind(new(repository.AdminRepository), new(*repository.AdminRepositoryImpl)),
	repository.ClientRepositoryInit, wire.Bind(new(repository.ClientRepository), new(*repository.ClientRepositoryImpl)),
	repository.CompanyRepositoryInit, wire.Bind(new(repository.CompanyRepository), new(*repository.CompanyRepositoryImpl)),
	repository.ParentMenuRepositoryInit, wire.Bind(new(repository.ParentMenuRepository), new(*repository.ParentMenuRepositoryImpl)),
	repository.MenuRepositoryInit, wire.Bind(new(repository.MenuRepository), new(*repository.MenuRepositoryImpl)),
	repository.RoleRepositoryInit, wire.Bind(new(repository.RoleRepository), new(*repository.RoleRepositoryImpl)),
	repository.RoleMenuAccessRepositoryInit, wire.Bind(new(repository.RoleMenuAccessRepository), new(*repository.RoleMenuAccessRepositoryImpl)),
	repository.CompanyUserRepositoryInit, wire.Bind(new(repository.CompanyUserRepository), new(*repository.CompanyUserRepositoryImpl)),
	repository.ClientUserRepositoryInit, wire.Bind(new(repository.ClientUserRepository), new(*repository.ClientUserRepositoryImpl)),
	repository.AuthRepositoryInit, wire.Bind(new(repository.AuthRepository), new(*repository.AuthRepositoryImpl)),
	repository.RateLimitRepositoryInit, wire.Bind(new(repository.RateLimitRepository), new(*repository.RateLimitRepositoryImpl)),
	repository.ClientSubscriptionRepositoryInit, wire.Bind(new(repository.ClientSubscriptionRepository), new(*repository.ClientSubscriptionRepositoryImpl)),
	repository.SubscriptionRepositoryInit, wire.Bind(new(repository.SubscriptionRepository), new(*repository.SubscriptionRepositoryImpl)),
	repository.ImageRepositoryInit, wire.Bind(new(repository.ImageRepository), new(*repository.ImageRepositoryImpl)),
	repository.ClientBranchRepositoryInit, wire.Bind(new(repository.ClientBranchRepository), new(*repository.ClientBranchRepositoryImpl)),
	repository.UserRepositoryInit, wire.Bind(new(repository.UserRepository), new(*repository.UserRepositoryImpl)),
	repository.ProductRepositoryInit, wire.Bind(new(repository.ProductRepository), new(*repository.ProductRepositoryImpl)),
	repository.StockTransferRepositoryInit, wire.Bind(new(repository.StockTransferRepository), new(*repository.StockTransferRepositoryImpl)),
	repository.POSTransactionRepositoryInit, wire.Bind(new(repository.POSTransactionRepository), new(*repository.POSTransactionRepositoryImpl)),
)

var serviceSet = wire.NewSet(
	service.NewAdminService, wire.Bind(new(service.AdminService), new(*service.AdminServiceImpl)),
	service.NewClientService, wire.Bind(new(service.ClientService), new(*service.ClientServiceImpl)),
	service.NewCompanyService, wire.Bind(new(service.CompanyService), new(*service.CompanyServiceImpl)),
	service.NewParentMenuService, wire.Bind(new(service.ParentMenuService), new(*service.ParentMenuServiceImpl)),
	service.NewMenuService, wire.Bind(new(service.MenuService), new(*service.MenuServiceImpl)),
	service.NewRoleService, wire.Bind(new(service.RoleService), new(*service.RoleServiceImpl)),
	service.NewRoleMenuAccessService, wire.Bind(new(service.RoleMenuAccessService), new(*service.RoleMenuAccessServiceImpl)),
	service.NewAuthService, wire.Bind(new(service.AuthService), new(*service.AuthServiceImpl)),
	service.NewSubscriptionService, wire.Bind(new(service.SubscriptionService), new(*service.SubscriptionServiceImpl)),
	service.NewSubscriptionGuardService, wire.Bind(new(service.SubscriptionGuardService), new(*service.SubscriptionGuardServiceImpl)),
	service.NewFileService, wire.Bind(new(service.FileService), new(*service.FileServiceImpl)),
	service.NewClientBranchService, wire.Bind(new(service.ClientBranchService), new(*service.ClientBranchServiceImpl)),
	service.NewUserService, wire.Bind(new(service.UserService), new(*service.UserServiceImpl)),
	service.NewClientUserService, wire.Bind(new(service.ClientUserService), new(*service.ClientUserServiceImpl)),
	service.NewProductService, wire.Bind(new(service.ProductService), new(*service.ProductServiceImpl)),
	service.NewStockTransferService, wire.Bind(new(service.StockTransferService), new(*service.StockTransferServiceImpl)),
	service.NewPOSTransactionService, wire.Bind(new(service.POSTransactionService), new(*service.POSTransactionServiceImpl)),
)

var controllerSet = wire.NewSet(
	controller.AdminControllerInit, wire.Bind(new(controller.AdminController), new(*controller.AdminControllerImpl)),
	controller.ClientControllerInit, wire.Bind(new(controller.ClientController), new(*controller.ClientControllerImpl)),
	controller.CompanyControllerInit, wire.Bind(new(controller.CompanyController), new(*controller.CompanyControllerImpl)),
	controller.ParentMenuControllerInit, wire.Bind(new(controller.ParentMenuController), new(*controller.ParentMenuControllerImpl)),
	controller.MenuControllerInit, wire.Bind(new(controller.MenuController), new(*controller.MenuControllerImpl)),
	controller.RoleControllerInit, wire.Bind(new(controller.RoleController), new(*controller.RoleControllerImpl)),
	controller.RoleAccessMenuControllerInit, wire.Bind(new(controller.RoleAccessMenuController), new(*controller.RoleAccessMenuControllerImpl)),
	controller.AuthControllerInit, wire.Bind(new(controller.AuthController), new(*controller.AuthControllerImpl)),
	controller.SubscriptionControllerInit, wire.Bind(new(controller.SubscriptionController), new(*controller.SubscriptionControllerImpl)),
	controller.FileControllerInit, wire.Bind(new(controller.FileController), new(*controller.FileControllerImpl)),
	controller.ClientBranchControllerInit, wire.Bind(new(controller.ClientBranchController), new(*controller.ClientBranchControllerImpl)),
	controller.UserControllerInit, wire.Bind(new(controller.UserController), new(*controller.UserControllerImpl)),
	controller.ClientUserControllerInit, wire.Bind(new(controller.ClientUserController), new(*controller.ClientUserControllerImpl)),
	controller.ProductControllerInit, wire.Bind(new(controller.ProductController), new(*controller.ProductControllerImpl)),
	controller.StockTransferControllerInit, wire.Bind(new(controller.StockTransferController), new(*controller.StockTransferControllerImpl)),
	controller.POSTransactionControllerInit, wire.Bind(new(controller.POSTransactionController), new(*controller.POSTransactionControllerImpl)),
)

func Init() *Initialization {
	wire.Build(
		NewInitialization,
		db,
		r2Set,
		repoSet,
		serviceSet,
		controllerSet,
	)
	return nil
}

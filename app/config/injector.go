//go:build wireinject
// +build wireinject

package config

import (
	"github.com/google/wire"

	"harjonan.id/user-service/app/controller"
	"harjonan.id/user-service/app/repository"
	"harjonan.id/user-service/app/service"
)

var db = wire.NewSet(ConnectToMongoDB)

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
)

func Init() *Initialization {
	wire.Build(
		NewInitialization,
		db,
		repoSet,
		serviceSet,
		controllerSet,
	)
	return nil
}

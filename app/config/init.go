package config

import (
	"harjonan.id/user-service/app/controller"
	"harjonan.id/user-service/app/repository"
	"harjonan.id/user-service/app/service"
)

type Initialization struct {
	AdminRepo          repository.AdminRepository
	ClientRepo         repository.ClientRepository
	CompanyRepo        repository.CompanyRepository
	ParentMenuRepo     repository.ParentMenuRepository
	MenuRepo           repository.MenuRepository
	RoleRepo           repository.RoleRepository
	RoleAccessMenuRepo repository.RoleMenuAccessRepository

	CompanyUserRepo        repository.CompanyUserRepository
	ClientUserRepo         repository.ClientUserRepository
	AuthRepo               repository.AuthRepository
	CheckRateLimitRepo     repository.RateLimitRepository
	ClientSubscriptionRepo repository.ClientSubscriptionRepository
	SubscriptionRepo       repository.SubscriptionRepository

	AdminSvc          service.AdminService
	ClientSvc         service.ClientService
	CompanySvc        service.CompanyService
	ParentMenuSvc     service.ParentMenuService
	MenuSvc           service.MenuService
	RoleSvc           service.RoleService
	RoleAccessMenuSvc service.RoleMenuAccessService
	AuthSvc           service.AuthService
	SubscriptionSvc   service.SubscriptionService

	AdminCtrl          controller.AdminController
	ClientCtrl         controller.ClientController
	CompanyCtrl        controller.CompanyController
	ParentMenuCtrl     controller.ParentMenuController
	MenuCtrl           controller.MenuController
	RoleCtrl           controller.RoleController
	RoleAccessMenuCtrl controller.RoleAccessMenuController
	AuthCtrl           controller.AuthController
	SubscriptionCtrl   controller.SubscriptionController
}

func NewInitialization(
	adminRepo repository.AdminRepository,
	clientRepo repository.ClientRepository,
	companyRepo repository.CompanyRepository,
	parentMenuRepo repository.ParentMenuRepository,
	menuRepo repository.MenuRepository,
	roleRepo repository.RoleRepository,
	roleAccessMenuRepo repository.RoleMenuAccessRepository,
	companyUserRepo repository.CompanyUserRepository,
	clientUserRepo repository.ClientUserRepository,
	authRepo repository.AuthRepository,
	checkRateLimitRepo repository.RateLimitRepository,
	clientSubscriptionRepo repository.ClientSubscriptionRepository,
	subscriptionRepo repository.SubscriptionRepository,

	adminSvc service.AdminService,
	clientSvc service.ClientService,
	companySvc service.CompanyService,
	parentMenuSvc service.ParentMenuService,
	menuSvc service.MenuService,
	roleSvc service.RoleService,
	roleAccessMenuSvc service.RoleMenuAccessService,
	authSvc service.AuthService,
	subscriptionSvc service.SubscriptionService,

	adminCtrl controller.AdminController,
	clientCtrl controller.ClientController,
	companyCtrl controller.CompanyController,
	parentMenuCtrl controller.ParentMenuController,
	menuCtrl controller.MenuController,
	roleCtrl controller.RoleController,
	roleAccessMenuCtrl controller.RoleAccessMenuController,
	authCtrl controller.AuthController,
	subscriptionCtrl controller.SubscriptionController,
) *Initialization {
	return &Initialization{
		AdminRepo:              adminRepo,
		ClientRepo:             clientRepo,
		CompanyRepo:            companyRepo,
		ParentMenuRepo:         parentMenuRepo,
		MenuRepo:               menuRepo,
		RoleRepo:               roleRepo,
		RoleAccessMenuRepo:     roleAccessMenuRepo,
		CompanyUserRepo:        companyUserRepo,
		ClientUserRepo:         clientUserRepo,
		AuthRepo:               authRepo,
		CheckRateLimitRepo:     checkRateLimitRepo,
		ClientSubscriptionRepo: clientSubscriptionRepo,
		SubscriptionRepo:       subscriptionRepo,

		AdminSvc:          adminSvc,
		ClientSvc:         clientSvc,
		CompanySvc:        companySvc,
		ParentMenuSvc:     parentMenuSvc,
		MenuSvc:           menuSvc,
		RoleSvc:           roleSvc,
		RoleAccessMenuSvc: roleAccessMenuSvc,
		AuthSvc:           authSvc,
		SubscriptionSvc:   subscriptionSvc,

		AdminCtrl:          adminCtrl,
		ClientCtrl:         clientCtrl,
		CompanyCtrl:        companyCtrl,
		ParentMenuCtrl:     parentMenuCtrl,
		MenuCtrl:           menuCtrl,
		RoleCtrl:           roleCtrl,
		RoleAccessMenuCtrl: roleAccessMenuCtrl,
		AuthCtrl:           authCtrl,
		SubscriptionCtrl:   subscriptionCtrl,
	}
}

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
	FileRepo               repository.ImageRepository
	ClientBranchRepo       repository.ClientBranchRepository
	UserRepo               repository.UserRepository
	ProductRepo            repository.ProductRepository
	StockTransferRepo      repository.StockTransferRepository
	PosTransactionRepo     repository.POSTransactionRepository

	AdminSvc          service.AdminService
	ClientSvc         service.ClientService
	CompanySvc        service.CompanyService
	ParentMenuSvc     service.ParentMenuService
	MenuSvc           service.MenuService
	RoleSvc           service.RoleService
	RoleAccessMenuSvc service.RoleMenuAccessService
	AuthSvc           service.AuthService
	SubscriptionSvc   service.SubscriptionService
	FileSvc           service.FileService
	ClientBranchSvc   service.ClientBranchService
	UserSvc           service.UserService
	ClientUserSvc     service.ClientUserService
	ProductSvc        service.ProductService
	StockTransferSvc  service.StockTransferService
	PosTransactionSvc service.POSTransactionService

	AdminCtrl          controller.AdminController
	ClientCtrl         controller.ClientController
	CompanyCtrl        controller.CompanyController
	ParentMenuCtrl     controller.ParentMenuController
	MenuCtrl           controller.MenuController
	RoleCtrl           controller.RoleController
	RoleAccessMenuCtrl controller.RoleAccessMenuController
	AuthCtrl           controller.AuthController
	SubscriptionCtrl   controller.SubscriptionController
	FileCtrl           controller.FileController
	ClientBranchCtrl   controller.ClientBranchController
	UserCtrl           controller.UserController
	ClientUserCtrl     controller.ClientUserController
	ProductCtrl        controller.ProductController
	StockTransferCtrl  controller.StockTransferController
	PosTransactionCtrl controller.POSTransactionController
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
	fileRepo repository.ImageRepository,
	clientBranchRepo repository.ClientBranchRepository,
	userRepo repository.UserRepository,
	productRepo repository.ProductRepository,
	stockTransferRepo repository.StockTransferRepository,
	posTransactionRepo repository.POSTransactionRepository,

	adminSvc service.AdminService,
	clientSvc service.ClientService,
	companySvc service.CompanyService,
	parentMenuSvc service.ParentMenuService,
	menuSvc service.MenuService,
	roleSvc service.RoleService,
	roleAccessMenuSvc service.RoleMenuAccessService,
	authSvc service.AuthService,
	subscriptionSvc service.SubscriptionService,
	fileSvc service.FileService,
	clientBranchSvc service.ClientBranchService,
	userSvc service.UserService,
	clientUserSvc service.ClientUserService,
	productSvc service.ProductService,
	stockTransferSvc service.StockTransferService,
	posTransactionSvc service.POSTransactionService,

	adminCtrl controller.AdminController,
	clientCtrl controller.ClientController,
	companyCtrl controller.CompanyController,
	parentMenuCtrl controller.ParentMenuController,
	menuCtrl controller.MenuController,
	roleCtrl controller.RoleController,
	roleAccessMenuCtrl controller.RoleAccessMenuController,
	authCtrl controller.AuthController,
	subscriptionCtrl controller.SubscriptionController,
	fileCtrl controller.FileController,
	clientBranchCtrl controller.ClientBranchController,
	userCtrl controller.UserController,
	clientUserCtrl controller.ClientUserController,
	productCtrl controller.ProductController,
	stockTransferCtrl controller.StockTransferController,
	posTransactionCtrl controller.POSTransactionController,
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
		FileRepo:               fileRepo,
		ClientBranchRepo:       clientBranchRepo,
		UserRepo:               userRepo,
		ProductRepo:            productRepo,
		StockTransferRepo:      stockTransferRepo,
		PosTransactionRepo:     posTransactionRepo,

		AdminSvc:          adminSvc,
		ClientSvc:         clientSvc,
		CompanySvc:        companySvc,
		ParentMenuSvc:     parentMenuSvc,
		MenuSvc:           menuSvc,
		RoleSvc:           roleSvc,
		RoleAccessMenuSvc: roleAccessMenuSvc,
		AuthSvc:           authSvc,
		SubscriptionSvc:   subscriptionSvc,
		FileSvc:           fileSvc,
		ClientBranchSvc:   clientBranchSvc,
		UserSvc:           userSvc,
		ClientUserSvc:     clientUserSvc,
		ProductSvc:        productSvc,
		StockTransferSvc:  stockTransferSvc,
		PosTransactionSvc: posTransactionSvc,

		AdminCtrl:          adminCtrl,
		ClientCtrl:         clientCtrl,
		CompanyCtrl:        companyCtrl,
		ParentMenuCtrl:     parentMenuCtrl,
		MenuCtrl:           menuCtrl,
		RoleCtrl:           roleCtrl,
		RoleAccessMenuCtrl: roleAccessMenuCtrl,
		AuthCtrl:           authCtrl,
		SubscriptionCtrl:   subscriptionCtrl,
		FileCtrl:           fileCtrl,
		ClientBranchCtrl:   clientBranchCtrl,
		UserCtrl:           userCtrl,
		ClientUserCtrl:     clientUserCtrl,
		ProductCtrl:        productCtrl,
		StockTransferCtrl:  stockTransferCtrl,
		PosTransactionCtrl: posTransactionCtrl,
	}
}

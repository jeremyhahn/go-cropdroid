package service

// import (
// 	"sync"

// 	"github.com/jeremyhahn/go-cropdroid/provisioner"
// 	"github.com/jeremyhahn/go-cropdroid/service"
// 	"github.com/jeremyhahn/go-cropdroid/shoppingcart"
// 	"github.com/op/go-logging"
// 	"github.com/stretchr/testify/mock"
// )

// type MockServiceRegistry struct {
// 	logger              *logging.Logger
// 	deviceServices      map[uint64][]service.DeviceServicer
// 	deviceServicesMutex *sync.RWMutex
// 	service.ServiceRegistry
// 	mock.Mock
// }

// func NewMockServiceRegistry(logger *logging.Logger) *MockServiceRegistry {
// 	return &MockServiceRegistry{
// 		logger: logger,
// 	}
// }

// func (registry *MockServiceRegistry) SetAlgorithmService(service service.AlgorithmServicer) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetAlgorithmService() service.AlgorithmServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetAuthService(service service.AuthServicer) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetAuthService() service.AuthServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetChannelService(service service.ChannelServicer) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetChannelService() service.ChannelServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetConditionService(service service.ConditionServicer) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetConditionService() service.ConditionServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetDeviceFactory(service service.DeviceFactory) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetDeviceFactory() service.DeviceFactory {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetDeviceServices(farmID uint64, deviceServices []service.DeviceServicer) {
// 	registry.deviceServicesMutex.Lock()
// 	defer registry.deviceServicesMutex.Unlock()
// 	registry.deviceServices[farmID] = deviceServices
// }

// func (registry *MockServiceRegistry) GetDeviceServices(farmID uint64) ([]service.DeviceServicer, error) {
// 	registry.deviceServicesMutex.Lock()
// 	defer registry.deviceServicesMutex.Unlock()
// 	if services, ok := registry.deviceServices[farmID]; ok {
// 		return services, nil
// 	}
// 	return nil, service.ErrFarmNotFound
// }

// func (registry *MockServiceRegistry) GetDeviceService(farmID uint64, deviceType string) (service.DeviceServicer, error) {
// 	return new(MockDeviceService), nil
// }

// func (registry *MockServiceRegistry) GetDeviceServiceByID(farmID uint64, deviceID uint64) (service.DeviceServicer, error) {
// 	registry.deviceServicesMutex.Lock()
// 	defer registry.deviceServicesMutex.Unlock()
// 	if services, ok := registry.deviceServices[farmID]; ok {
// 		for _, service := range services {
// 			if service.ID() == deviceID {
// 				return service, nil
// 			}
// 		}
// 		return nil, service.ErrDeviceNotFound
// 	}
// 	return nil, service.ErrFarmNotFound
// }

// func (registry *MockServiceRegistry) SetDeviceService(farmID uint64, deviceService service.DeviceServicer) (service.DeviceServicer, error) {
// 	registry.logger.Fatal("not implemented")
// 	return nil, nil
// }

// func (registry *MockServiceRegistry) AddEventLogService(eventLogService service.EventLogServicer) error {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetEventLogService(eventLogServices map[uint64]service.EventLogServicer) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetEventLogServices() map[uint64]service.EventLogServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) GetEventLogService(farmID uint64) service.EventLogServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) RemoveEventLogService(farmID uint64) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) SetFarmFactory(FarmFactory service.FarmFactory) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetFarmFactory() service.FarmFactory {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) AddFarmService(farmService service.FarmServicer) error {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetFarmServices(map[uint64]service.FarmServicer) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetFarmServices() map[uint64]service.FarmServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) GetFarmService(uint64) service.FarmServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) RemoveFarmService(farmID uint64) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) SetFarmProvisioner(farmProvisioner provisioner.FarmProvisioner) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetFarmProvisioner() provisioner.FarmProvisioner {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetGoogleAuthService(googleAuthService service.AuthServicer) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetGoogleAuthService() service.AuthServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetMetricService(service service.MetricService) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetMetricService() service.MetricService {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetNotificationService(service.NotificationServicer) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetNotificationService() service.NotificationServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetScheduleService(service.ScheduleService) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetScheduleService() service.ScheduleService {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetShoppingCartService(shoppingcart.ShoppingCartService) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetShoppingCartService() shoppingcart.ShoppingCartService {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetOrganizationService(organizationService service.OrganizationService) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetOrganizationService() service.OrganizationService {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetRoleService(roleService service.RoleServicer) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetRoleService() service.RoleServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetUserService(service.UserServicer) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetUserService() service.UserServicer {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetWorkflowService(service.WorkflowService) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetWorkflowService() service.WorkflowService {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

// func (registry *MockServiceRegistry) SetWorkflowStepService(service.WorkflowStepService) {
// 	registry.logger.Fatal("not implemented")
// }

// func (registry *MockServiceRegistry) GetWorkflowStepService() service.WorkflowStepService {
// 	registry.logger.Fatal("not implemented")
// 	return nil
// }

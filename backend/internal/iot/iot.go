package iot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/google/uuid"
)

// =============================================================================
// Types
// =============================================================================

// DeviceType represents the type of IoT device
type DeviceType string

const (
	DeviceTypeSensor    DeviceType = "sensor"
	DeviceTypeActuator  DeviceType = "actuator"
	DeviceTypeController DeviceType = "controller"
	DeviceTypeGateway   DeviceType = "gateway"
	DeviceTypeCamera    DeviceType = "camera"
	DeviceTypeDisplay   DeviceType = "display"
)

// DeviceStatus represents the status of an IoT device
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusError   DeviceStatus = "error"
	DeviceStatusPending DeviceStatus = "pending"
)

// Device represents an IoT device
type Device struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	BusinessID  uuid.UUID              `json:"business_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        DeviceType             `json:"type"`
	Status      DeviceStatus           `json:"status"`
	Location    string                 `json:"location"`
	Metadata    map[string]interface{} `json:"metadata"`
	LastSeen    time.Time              `json:"last_seen"`
	LastData    map[string]interface{} `json:"last_data,omitempty"`
	Config      DeviceConfig           `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// DeviceConfig contains configuration for a device
type DeviceConfig struct {
	PollInterval    int               `json:"poll_interval"` // seconds
	ReportInterval  int               `json:"report_interval"` // seconds
	Thresholds      map[string]Threshold `json:"thresholds,omitempty"`
	Actions         []Action          `json:"actions,omitempty"`
	Credentials     *Credentials      `json:"-"`
}

// Threshold defines alert thresholds for a metric
type Threshold struct {
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
	AlertOnBreach bool   `json:"alert_on_breach"`
}

// Action defines an automated action for a device
type Action struct {
	Trigger   string                 `json:"trigger"` // threshold_breach, schedule, manual
	Command   string                 `json:"command"`
	Params    map[string]interface{} `json:"params,omitempty"`
	AgentID   *uuid.UUID             `json:"agent_id,omitempty"`
}

// Credentials stores device authentication
type Credentials struct {
	Type     string `json:"type"` // api_key, oauth, certificate
	APIKey   string `json:"api_key,omitempty"`
	Token    string `json:"token,omitempty"`
	CertPath string `json:"cert_path,omitempty"`
}

// DataPoint represents a single data point from a device
type DataPoint struct {
	DeviceID  uuid.UUID              `json:"device_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Command represents a command to send to a device
type Command struct {
	ID        uuid.UUID              `json:"id"`
	DeviceID  uuid.UUID              `json:"device_id"`
	Action    string                 `json:"action"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Status    string                 `json:"status"` // pending, sent, acknowledged, completed, failed
	CreatedAt time.Time              `json:"created_at"`
	SentAt    *time.Time             `json:"sent_at,omitempty"`
	CompletedAt *time.Time           `json:"completed_at,omitempty"`
}

// DeviceGroup represents a group of devices
type DeviceGroup struct {
	ID          uuid.UUID   `json:"id"`
	TenantID    uuid.UUID   `json:"tenant_id"`
	BusinessID  uuid.UUID   `json:"business_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	DeviceIDs   []uuid.UUID `json:"device_ids"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// =============================================================================
// Service
// =============================================================================

// Service handles IoT device management
type Service struct {
	log           *logger.Logger
	devices       map[uuid.UUID]*Device
	devicesMu     sync.RWMutex
	dataBuffer    chan DataPoint
	commandQueue  chan Command
	adapters      map[string]Adapter
}

// Adapter interface for IoT protocols
type Adapter interface {
	// Connect establishes connection to the device
	Connect(ctx context.Context, device *Device) error
	
	// Disconnect closes connection to the device
	Disconnect(ctx context.Context, device *Device) error
	
	// ReadData reads data from the device
	ReadData(ctx context.Context, device *Device) (map[string]interface{}, error)
	
	// SendCommand sends a command to the device
	SendCommand(ctx context.Context, device *Device, cmd *Command) error
	
	// Subscribe subscribes to device data updates
	Subscribe(ctx context.Context, device *Device, handler func(DataPoint)) error
}

// NewService creates a new IoT service
func NewService(log *logger.Logger) *Service {
	s := &Service{
		log:          log,
		devices:      make(map[uuid.UUID]*Device),
		dataBuffer:   make(chan DataPoint, 10000),
		commandQueue: make(chan Command, 1000),
		adapters:     make(map[string]Adapter),
	}

	// Start background workers
	go s.processDataBuffer()
	go s.processCommandQueue()

	return s
}

// RegisterAdapter registers an IoT protocol adapter
func (s *Service) RegisterAdapter(protocol string, adapter Adapter) {
	s.adapters[protocol] = adapter
}

// =============================================================================
// Device Management
// =============================================================================

// RegisterDevice registers a new IoT device
func (s *Service) RegisterDevice(ctx context.Context, device *Device) error {
	device.ID = uuid.New()
	device.Status = DeviceStatusPending
	device.CreatedAt = time.Now()
	device.UpdatedAt = time.Now()

	s.devicesMu.Lock()
	s.devices[device.ID] = device
	s.devicesMu.Unlock()

	s.log.Infow("device registered",
		"device_id", device.ID,
		"name", device.Name,
		"type", device.Type,
	)

	return nil
}

// GetDevice retrieves a device by ID
func (s *Service) GetDevice(ctx context.Context, deviceID uuid.UUID) (*Device, error) {
	s.devicesMu.RLock()
	device, ok := s.devices[deviceID]
	s.devicesMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	return device, nil
}

// UpdateDevice updates a device
func (s *Service) UpdateDevice(ctx context.Context, device *Device) error {
	s.devicesMu.Lock()
	if _, ok := s.devices[device.ID]; !ok {
		s.devicesMu.Unlock()
		return fmt.Errorf("device not found: %s", device.ID)
	}
	device.UpdatedAt = time.Now()
	s.devices[device.ID] = device
	s.devicesMu.Unlock()

	return nil
}

// DeleteDevice removes a device
func (s *Service) DeleteDevice(ctx context.Context, deviceID uuid.UUID) error {
	s.devicesMu.Lock()
	delete(s.devices, deviceID)
	s.devicesMu.Unlock()

	s.log.Infow("device deleted", "device_id", deviceID)
	return nil
}

// ListDevices lists devices with optional filters
func (s *Service) ListDevices(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}) ([]*Device, error) {
	s.devicesMu.RLock()
	defer s.devicesMu.RUnlock()

	var devices []*Device
	for _, device := range s.devices {
		if device.TenantID != tenantID {
			continue
		}

		// Apply filters
		if deviceType, ok := filters["type"].(DeviceType); ok && device.Type != deviceType {
			continue
		}
		if status, ok := filters["status"].(DeviceStatus); ok && device.Status != status {
			continue
		}
		if businessID, ok := filters["business_id"].(uuid.UUID); ok && device.BusinessID != businessID {
			continue
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// =============================================================================
// Data Collection
// =============================================================================

// IngestData receives data from a device
func (s *Service) IngestData(ctx context.Context, data DataPoint) error {
	select {
	case s.dataBuffer <- data:
		// Update device last seen and last data
		s.devicesMu.Lock()
		if device, ok := s.devices[data.DeviceID]; ok {
			device.LastSeen = data.Timestamp
			device.LastData = data.Data
			device.Status = DeviceStatusOnline
		}
		s.devicesMu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("data buffer full")
	}
}

// processDataBuffer processes incoming device data
func (s *Service) processDataBuffer() {
	for data := range s.dataBuffer {
		// Check thresholds and trigger alerts
		s.devicesMu.RLock()
		device, ok := s.devices[data.DeviceID]
		s.devicesMu.RUnlock()

		if !ok {
			continue
		}

		for metric, threshold := range device.Config.Thresholds {
			value, ok := data.Data[metric].(float64)
			if !ok {
				continue
			}

			if threshold.Min != nil && value < *threshold.Min {
				s.handleThresholdBreach(device, metric, value, "below_min")
			}
			if threshold.Max != nil && value > *threshold.Max {
				s.handleThresholdBreach(device, metric, value, "above_max")
			}
		}

		// Store data point (would integrate with time series database)
		s.log.Debugw("data point processed",
			"device_id", data.DeviceID,
			"data", data.Data,
		)
	}
}

// handleThresholdBreach handles threshold breach events
func (s *Service) handleThresholdBreach(device *Device, metric string, value float64, breachType string) {
	s.log.Warnw("threshold breach",
		"device_id", device.ID,
		"device_name", device.Name,
		"metric", metric,
		"value", value,
		"breach_type", breachType,
	)

	// Trigger configured actions
	for _, action := range device.Config.Actions {
		if action.Trigger == "threshold_breach" {
			cmd := Command{
				ID:        uuid.New(),
				DeviceID:  device.ID,
				Action:    action.Command,
				Params:    action.Params,
				Status:    "pending",
				CreatedAt: time.Now(),
			}
			s.commandQueue <- cmd
		}
	}
}

// =============================================================================
// Command Execution
// =============================================================================

// SendCommand sends a command to a device
func (s *Service) SendCommand(ctx context.Context, cmd *Command) error {
	cmd.ID = uuid.New()
	cmd.Status = "pending"
	cmd.CreatedAt = time.Now()

	select {
	case s.commandQueue <- *cmd:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("command queue full")
	}
}

// processCommandQueue processes outgoing device commands
func (s *Service) processCommandQueue() {
	for cmd := range s.commandQueue {
		s.devicesMu.RLock()
		device, ok := s.devices[cmd.DeviceID]
		s.devicesMu.RUnlock()

		if !ok {
			s.log.Warnw("command for unknown device",
				"command_id", cmd.ID,
				"device_id", cmd.DeviceID,
			)
			continue
		}

		// Get adapter for device protocol
		protocol := "mqtt" // Default protocol
		if p, ok := device.Metadata["protocol"].(string); ok {
			protocol = p
		}

		adapter, ok := s.adapters[protocol]
		if !ok {
			s.log.Warnw("no adapter for protocol",
				"protocol", protocol,
				"device_id", device.ID,
			)
			continue
		}

		// Send command
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := adapter.SendCommand(ctx, device, &cmd); err != nil {
			s.log.Errorw("failed to send command",
				"command_id", cmd.ID,
				"device_id", device.ID,
				"error", err,
			)
			cmd.Status = "failed"
		} else {
			now := time.Now()
			cmd.SentAt = &now
			cmd.Status = "sent"
			s.log.Infow("command sent",
				"command_id", cmd.ID,
				"device_id", device.ID,
				"action", cmd.Action,
			)
		}
		cancel()
	}
}

// =============================================================================
// Device Groups
// =============================================================================

// CreateDeviceGroup creates a new device group
func (s *Service) CreateDeviceGroup(ctx context.Context, group *DeviceGroup) error {
	group.ID = uuid.New()
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()

	s.log.Infow("device group created",
		"group_id", group.ID,
		"name", group.Name,
		"device_count", len(group.DeviceIDs),
	)

	return nil
}

// SendGroupCommand sends a command to all devices in a group
func (s *Service) SendGroupCommand(ctx context.Context, group *DeviceGroup, action string, params map[string]interface{}) error {
	for _, deviceID := range group.DeviceIDs {
		cmd := &Command{
			DeviceID: deviceID,
			Action:   action,
			Params:   params,
		}
		if err := s.SendCommand(ctx, cmd); err != nil {
			s.log.Warnw("failed to queue command for device",
				"device_id", deviceID,
				"group_id", group.ID,
				"error", err,
			)
		}
	}
	return nil
}

// =============================================================================
// Health Monitoring
// =============================================================================

// CheckDeviceHealth checks the health of all devices
func (s *Service) CheckDeviceHealth(ctx context.Context) error {
	s.devicesMu.Lock()
	defer s.devicesMu.Unlock()

	now := time.Now()
	offlineThreshold := 5 * time.Minute

	for _, device := range s.devices {
		if device.Status == DeviceStatusOnline && now.Sub(device.LastSeen) > offlineThreshold {
			device.Status = DeviceStatusOffline
			s.log.Warnw("device went offline",
				"device_id", device.ID,
				"name", device.Name,
				"last_seen", device.LastSeen,
			)
		}
	}

	return nil
}

// GetHealthSummary returns a summary of device health
func (s *Service) GetHealthSummary(ctx context.Context, tenantID uuid.UUID) map[string]int {
	s.devicesMu.RLock()
	defer s.devicesMu.RUnlock()

	summary := map[string]int{
		"online":  0,
		"offline": 0,
		"error":   0,
		"pending": 0,
		"total":   0,
	}

	for _, device := range s.devices {
		if device.TenantID != tenantID {
			continue
		}
		summary["total"]++
		summary[string(device.Status)]++
	}

	return summary
}


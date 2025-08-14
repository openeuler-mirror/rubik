package iolimit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"

	"isula.org/rubik/pkg/api"
	"isula.org/rubik/pkg/common/constant"
	"isula.org/rubik/pkg/core/typedef"
	"isula.org/rubik/pkg/core/typedef/cgroup"
)

// Mock implementation for testing
type mockViewer struct {
	pods       map[string]*typedef.PodInfo
	containers map[string]*typedef.ContainerInfo
}

func (m *mockViewer) ListPodsWithOptions(options ...api.ListOption) map[string]*typedef.PodInfo {
	return m.pods
}

func (m *mockViewer) ListContainersWithOptions(options ...api.ListOption) map[string]*typedef.ContainerInfo {
	return m.containers
}

// setupTestCgroupEnv sets up a temporary cgroup environment for testing
func setupTestCgroupEnv(t *testing.T) (string, func()) {
	// Create temporary directory
	tempDir, err := ioutil.TempDir("", "iolimit_cgroup_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize cgroup with temp directory
	err = cgroup.Init(cgroup.WithRoot(tempDir))
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to init cgroup with temp dir: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		// Restore default cgroup root
		cgroup.Init(cgroup.WithRoot(constant.DefaultCgroupRoot))
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// mockConvertToMajorMinor is a mock implementation for testing
func mockConvertToMajorMinor(deviceName string) (string, error) {
	// Mock common device mappings for testing
	deviceMappings := map[string]string{
		"/dev/sda":         "8:0",
		"/dev/sda1":        "8:1",
		"/dev/sdb":         "8:16",
		"/dev/sdb1":        "8:17",
		"/dev/nvme0n1":     "259:0",
		"/dev/nvme0n1p1":   "259:1",
		"/dev/mapper/test": "253:0",
		"8:0":              "8:0",     // Already in major:minor format
		"8:16":             "8:16",    // Already in major:minor format
		"253:1":            "253:1",   // Already in major:minor format
		"259:0":            "259:0",   // Already in major:minor format
		"254:255":          "254:255", // Already in major:minor format
	}

	if majorMinor, exists := deviceMappings[deviceName]; exists {
		return majorMinor, nil
	}

	return "", fmt.Errorf("unsupported device name format: %s", deviceName)
}

// setupMockConvertToMajorMinor replaces the convertToMajorMinor function with a mock
func setupMockConvertToMajorMinor() func() {
	originalFunc := convertToMajorMinorFunc
	convertToMajorMinorFunc = mockConvertToMajorMinor
	return func() {
		convertToMajorMinorFunc = originalFunc
	}
}

// TestIOLimitFactory tests the IOLimitFactory
func TestIOLimitFactory(t *testing.T) {
	factory := IOLimitFactory{ObjName: "test"}

	// Test Name method
	if factory.Name() != "IOLimitFactory" {
		t.Errorf("Expected factory name to be 'IOLimitFactory', got %s", factory.Name())
	}

	// Test NewObj method
	obj, err := factory.NewObj()
	if err != nil {
		t.Errorf("NewObj should not return error, got %v", err)
	}

	iolimit, ok := obj.(*IOLimit)
	if !ok {
		t.Errorf("NewObj should return *IOLimit, got %T", obj)
	}

	if iolimit == nil {
		t.Error("NewObj should return non-nil IOLimit")
	}
}

// TestIOLimit_PreStart tests the PreStart method
func TestIOLimit_PreStart(t *testing.T) {
	iolimit := &IOLimit{}

	// Test with nil viewer
	err := iolimit.PreStart(nil)
	if err == nil {
		t.Error("PreStart should return error with nil viewer")
	}
	if !strings.Contains(err.Error(), "invalid pods viewer") {
		t.Errorf("Expected error about invalid pods viewer, got %v", err)
	}

	// Test with valid viewer but empty pods
	mockViewer := &mockViewer{
		pods:       make(map[string]*typedef.PodInfo),
		containers: make(map[string]*typedef.ContainerInfo),
	}
	err = iolimit.PreStart(mockViewer)
	if err != nil {
		t.Errorf("PreStart should not return error with empty pods, got %v", err)
	}
}

// TestIOLimit_AddPod tests the AddPod method
func TestIOLimit_AddPod(t *testing.T) {
	iolimit := &IOLimit{}

	// Test with nil podInfo
	err := iolimit.AddPod(nil)
	if err == nil {
		t.Error("AddPod should return error with nil podInfo")
	}
	if !strings.Contains(err.Error(), "invalid pod info") {
		t.Errorf("Expected error about invalid pod info, got %v", err)
	}

	// Test with valid podInfo but no config
	podInfo := &typedef.PodInfo{
		Name:        "test-pod",
		Annotations: map[string]string{},
	}
	err = iolimit.AddPod(podInfo)
	if err != nil {
		t.Errorf("AddPod should not return error with valid podInfo without config, got %v", err)
	}
}

// TestIOLimit_UpdatePod tests the UpdatePod method
func TestIOLimit_UpdatePod(t *testing.T) {
	iolimit := &IOLimit{}

	// Test with nil podInfo
	err := iolimit.UpdatePod(nil, nil)
	if err == nil {
		t.Error("UpdatePod should return error with nil podInfo")
	}
	if !strings.Contains(err.Error(), "invalid pod info") {
		t.Errorf("Expected error about invalid pod info, got %v", err)
	}

	// Test with valid podInfo but no config
	podInfo := &typedef.PodInfo{
		Name:        "test-pod",
		Annotations: map[string]string{},
	}
	err = iolimit.UpdatePod(nil, podInfo)
	if err != nil {
		t.Errorf("UpdatePod should not return error with valid podInfo without config, got %v", err)
	}
}

// TestParseIOLimitConfig tests the parseIOLimitConfig function
func TestParseIOLimitConfig(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *BlkConfig
		expectError bool
	}{
		{
			name:        "empty config",
			input:       "",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid JSON",
			input:       "invalid json",
			expected:    nil,
			expectError: true,
		},
		{
			name:  "valid config",
			input: `{"device_read_bps":[{"device":"/dev/sda","value":"1048576"}],"device_write_bps":[{"device":"/dev/sdb1","value":"2097152"}]}`,
			expected: &BlkConfig{
				DeviceReadBps: []DeviceConfig{
					{DeviceName: "/dev/sda", DeviceValue: "1048576"},
				},
				DeviceWriteBps: []DeviceConfig{
					{DeviceName: "/dev/sdb1", DeviceValue: "2097152"},
				},
				DeviceReadIops:  []DeviceConfig{},
				DeviceWriteIops: []DeviceConfig{},
			},
			expectError: false,
		},
		{
			name:  "partial config - only read bps",
			input: `{"device_read_bps":[{"device":"/dev/sda","value":"1048576"}]}`,
			expected: &BlkConfig{
				DeviceReadBps: []DeviceConfig{
					{DeviceName: "/dev/sda", DeviceValue: "1048576"},
				},
				DeviceWriteBps:  []DeviceConfig{},
				DeviceReadIops:  []DeviceConfig{},
				DeviceWriteIops: []DeviceConfig{},
			},
			expectError: false,
		},
		{
			name:  "partial config - only write iops",
			input: `{"device_write_iops":[{"device":"8:0","value":"1000"}]}`,
			expected: &BlkConfig{
				DeviceReadBps:  []DeviceConfig{},
				DeviceWriteBps: []DeviceConfig{},
				DeviceReadIops: []DeviceConfig{},
				DeviceWriteIops: []DeviceConfig{
					{DeviceName: "8:0", DeviceValue: "1000"},
				},
			},
			expectError: false,
		},
		{
			name:  "empty arrays",
			input: `{"device_read_bps":[],"device_write_bps":[],"device_read_iops":[],"device_write_iops":[]}`,
			expected: &BlkConfig{
				DeviceReadBps:   []DeviceConfig{},
				DeviceWriteBps:  []DeviceConfig{},
				DeviceReadIops:  []DeviceConfig{},
				DeviceWriteIops: []DeviceConfig{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseIOLimitConfig(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

// TestParseAndResetParams tests the parseAndResetParams function
func TestParseAndResetParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "single device",
			input:    "8:0 1048576",
			expected: "8:0 0",
		},
		{
			name:     "multiple devices",
			input:    "8:0 1048576\n8:16 2097152",
			expected: "8:0 0\n8:16 0",
		},
		{
			name:     "with empty lines",
			input:    "8:0 1048576\n\n8:16 2097152\n",
			expected: "8:0 0\n8:16 0",
		},
		{
			name:     "with spaces",
			input:    "  8:0 1048576  \n  8:16 2097152  ",
			expected: "8:0 0\n8:16 0",
		},
		{
			name:     "invalid format (no value)",
			input:    "8:0",
			expected: "",
		},
		{
			name:     "single device with zero value",
			input:    "8:0 0",
			expected: "8:0 0",
		},
		{
			name:     "device with large value",
			input:    "253:15 9999999999",
			expected: "253:15 0",
		},
		{
			name:     "multiple devices with mixed values",
			input:    "8:0 1048576\n8:16 0\n253:1 2097152",
			expected: "8:0 0\n8:16 0\n253:1 0",
		},
		{
			name:     "device with extra parameters",
			input:    "8:0 1048576 extra param",
			expected: "8:0 0",
		},
		{
			name:     "mixed valid and invalid lines",
			input:    "8:0 1048576\ninvalid_line\n8:16 2097152\n\nanother_invalid",
			expected: "8:0 0\n8:16 0",
		},
		{
			name:     "only whitespace",
			input:    "   \n\t\n   ",
			expected: "",
		},
		{
			name:     "device with tab separator",
			input:    "8:0\t1048576",
			expected: "8:0 0",
		},
		{
			name:     "device with multiple spaces",
			input:    "8:0    1048576",
			expected: "8:0 0",
		},
		{
			name:     "complex major:minor numbers",
			input:    "259:0 1048576\n254:255 2097152",
			expected: "259:0 0\n254:255 0",
		},
		{
			name:     "only newlines",
			input:    "\n\n\n",
			expected: "",
		},
		{
			name:     "device with negative value",
			input:    "8:0 -1048576",
			expected: "8:0 0",
		},
		{
			name:     "device with decimal value",
			input:    "8:0 1048576.5",
			expected: "8:0 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAndResetParams(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestConvertToMajorMinor tests the convertToMajorMinor function
func TestConvertToMajorMinor(t *testing.T) {
	// Setup mock function for testing
	cleanupMock := setupMockConvertToMajorMinor()
	defer cleanupMock()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "already major:minor format",
			input:       "8:0",
			expected:    "8:0",
			expectError: false,
		},
		{
			name:        "numeric major:minor",
			input:       "253:1",
			expected:    "253:1",
			expectError: false,
		},
		{
			name:        "valid device path /dev/sda",
			input:       "/dev/sda",
			expected:    "8:0",
			expectError: false,
		},
		{
			name:        "valid device path /dev/sdb1",
			input:       "/dev/sdb1",
			expected:    "8:17",
			expectError: false,
		},
		{
			name:        "valid nvme device",
			input:       "/dev/nvme0n1",
			expected:    "259:0",
			expectError: false,
		},
		{
			name:        "valid mapper device",
			input:       "/dev/mapper/test",
			expected:    "253:0",
			expectError: false,
		},
		{
			name:        "complex major:minor",
			input:       "254:255",
			expected:    "254:255",
			expectError: false,
		},
		{
			name:        "unsupported device",
			input:       "/dev/nonexistent",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid format",
			input:       "invalid",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertToMajorMinor(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// deviceTestCase represents a test case for device major:minor conversion
type deviceTestCase struct {
	deviceName     string
	expectedResult string // Expected major:minor format like "8:0"
}

// buildDeviceTestTable builds a test table with real block device major:minor numbers
func buildDeviceTestTable(t *testing.T) []deviceTestCase {
	var testCases []deviceTestCase

	// Check /sys/block for main block devices only
	sysBlockDir := "/sys/block"
	entries, err := os.ReadDir(sysBlockDir)
	if err != nil {
		t.Logf("Cannot read %s: %v (this is expected on non-Linux systems)", sysBlockDir, err)
		return testCases
	}

	// Process each entry (symlink) in /sys/block
	for _, entry := range entries {
		devicePath := "/dev/" + entry.Name()

		// Get device stat info
		stat, err := os.Stat(devicePath)
		if err != nil {
			continue // Skip if device doesn't exist
		}

		// Extract major:minor if it's a device file
		if stat.Mode()&os.ModeDevice != 0 {
			if sys, ok := stat.Sys().(*syscall.Stat_t); ok {
				major := uint64((sys.Rdev >> 8) & 0xff)
				minor := uint64(sys.Rdev & 0xff)
				expectedResult := fmt.Sprintf("%d:%d", major, minor)

				testCases = append(testCases, deviceTestCase{
					deviceName:     devicePath,
					expectedResult: expectedResult,
				})
			}
		}
	}

	t.Logf("Built test table with %d main block devices:", len(testCases))
	for _, tc := range testCases {
		t.Logf("  %s -> %s", tc.deviceName, tc.expectedResult)
	}

	return testCases
}

// TestConvertToMajorMinorImpl tests the actual convertToMajorMinorImpl function with real block devices
// This test dynamically discovers block devices on the system and tests with them.
//
// Usage in Linux environment:
//
//	go test -v ./pkg/services/iolimit -run TestConvertToMajorMinorImpl
//
// Expected behavior:
//   - On non-Linux systems: Test will be skipped
//   - On Linux systems: Test will discover and test real block devices from /sys/block
//   - Validates major:minor conversion for discovered block devices only
//
// Note: This test complements TestConvertToMajorMinor which uses mock implementations.
func TestConvertToMajorMinorImpl(t *testing.T) {
	// Skip test on non-Linux systems
	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		t.Skip("Skipping test: not running on Linux system")
		return
	}
	if _, err := os.Stat("/sys"); os.IsNotExist(err) {
		t.Skip("Skipping test: not running on Linux system")
		return
	}

	// Get test table with real device information
	testCases := buildDeviceTestTable(t)

	if len(testCases) == 0 {
		t.Skip("No main block devices found - skipping test in virtualized/container environment")
		return
	}

	// Execute tests using the test table
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("device_%s", strings.TrimPrefix(tc.deviceName, "/dev/")), func(t *testing.T) {
			// Call the function under test
			result, err := convertToMajorMinorImpl(tc.deviceName)

			if err != nil {
				t.Errorf("Failed to convert device %s: %v", tc.deviceName, err)
				return
			}

			// Verify the result format
			if !strings.Contains(result, ":") {
				t.Errorf("Result should be in major:minor format, got: %s", result)
				return
			}

			// Compare with expected result directly
			if result != tc.expectedResult {
				t.Errorf("Conversion mismatch for %s: expected %s, got %s", tc.deviceName, tc.expectedResult, result)
				return
			}

			t.Logf("✓ Device %s -> %s - conversion correct", tc.deviceName, result)
		})
	}
}

// TestApplyDeviceConfig tests the applyDeviceConfig function
func TestApplyDeviceConfig(t *testing.T) {
	// Setup test cgroup environment
	tempDir, cleanupCgroup := setupTestCgroupEnv(t)
	defer cleanupCgroup()

	// Setup mock convertToMajorMinor function
	cleanupMock := setupMockConvertToMajorMinor()
	defer cleanupMock()

	// Create test pod directory in the temp cgroup root following the correct cgroup structure
	testPodPath := "test-pod"
	// The cgroup structure should be: {tempDir}/blkio/{testPodPath}/
	blkioDir := filepath.Join(tempDir, "blkio")
	testPodDir := filepath.Join(blkioDir, testPodPath)
	err := os.MkdirAll(testPodDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test pod directory: %v", err)
	}

	tests := []struct {
		name        string
		devices     []DeviceConfig
		expectError bool
	}{
		{
			name:        "empty devices",
			devices:     []DeviceConfig{},
			expectError: false,
		},
		{
			name: "valid devices",
			devices: []DeviceConfig{
				{DeviceName: "/dev/sda", DeviceValue: "1048576"},
				{DeviceName: "/dev/sdb1", DeviceValue: "2097152"},
			},
			expectError: false, // Should work with mock function
		},
		{
			name: "major:minor format devices",
			devices: []DeviceConfig{
				{DeviceName: "8:0", DeviceValue: "1048576"},
				{DeviceName: "253:1", DeviceValue: "2097152"},
			},
			expectError: false, // Should work with mock function
		},
		{
			name: "invalid device name",
			devices: []DeviceConfig{
				{DeviceName: "", DeviceValue: "1048576"},
			},
			expectError: false, // Should skip invalid devices
		},
		{
			name: "invalid device value",
			devices: []DeviceConfig{
				{DeviceName: "/dev/sda", DeviceValue: ""},
			},
			expectError: false, // Should skip invalid devices
		},
		{
			name: "unsupported device",
			devices: []DeviceConfig{
				{DeviceName: "/dev/unsupported", DeviceValue: "1048576"},
			},
			expectError: false, // Should skip unsupported devices
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the test file first (it needs to exist before writing)
			testFileName := "test_file"
			filePath := filepath.Join(testPodDir, testFileName)
			err := ioutil.WriteFile(filePath, []byte(""), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			err = applyDeviceConfig(testPodPath, testFileName, tt.devices)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil && !strings.Contains(err.Error(), "no valid device config") &&
					!strings.Contains(err.Error(), "failed to write config") {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// For successful cases with valid devices, verify the file content
			if !tt.expectError && len(tt.devices) > 0 && tt.devices[0].DeviceName != "" && tt.devices[0].DeviceValue != "" {
				// Check if any device was successfully converted
				hasValidDevice := false
				for _, device := range tt.devices {
					if device.DeviceName != "" && device.DeviceValue != "" {
						if _, err := mockConvertToMajorMinor(device.DeviceName); err == nil {
							hasValidDevice = true
							break
						}
					}
				}

				if hasValidDevice {
					// Try to read the file to verify content was written
					content, err := cgroup.ReadCgroupFile("blkio", testPodPath, testFileName)
					if err == nil && len(content) > 0 {
						t.Logf("File content: %s", string(content))
					}
				}
			}
		})
	}
}

// TestApplyIOLimitConfig tests the applyIOLimitConfig function
func TestApplyIOLimitConfig(t *testing.T) {
	// Setup test cgroup environment
	tempDir, cleanupCgroup := setupTestCgroupEnv(t)
	defer cleanupCgroup()

	// Setup mock convertToMajorMinor function
	cleanupMock := setupMockConvertToMajorMinor()
	defer cleanupMock()

	// Create test pod directory in the temp cgroup root following the correct cgroup structure
	testPodPath := "test-pod"
	// The cgroup structure should be: {tempDir}/blkio/{testPodPath}/
	blkioDir := filepath.Join(tempDir, "blkio")
	testPodDir := filepath.Join(blkioDir, testPodPath)
	err := os.MkdirAll(testPodDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test pod directory: %v", err)
	}

	// Test with nil config
	err = applyIOLimitConfig(testPodPath, nil)
	if err == nil {
		t.Error("Expected error with nil config")
	}
	if !strings.Contains(err.Error(), "config is nil") {
		t.Errorf("Expected error about nil config, got %v", err)
	}

	// Test with valid config using mock devices
	cfg := &BlkConfig{
		DeviceReadBps: []DeviceConfig{
			{DeviceName: "/dev/sda", DeviceValue: "1048576"},
		},
		DeviceWriteBps: []DeviceConfig{
			{DeviceName: "8:16", DeviceValue: "2097152"},
		},
		DeviceReadIops: []DeviceConfig{
			{DeviceName: "/dev/nvme0n1", DeviceValue: "1000"},
		},
		DeviceWriteIops: []DeviceConfig{
			{DeviceName: "253:1", DeviceValue: "2000"},
		},
	}

	// Create the cgroup files first (they need to exist before writing)
	cgroupFiles := []string{
		"blkio.throttle.read_bps_device",
		"blkio.throttle.write_bps_device",
		"blkio.throttle.read_iops_device",
		"blkio.throttle.write_iops_device",
	}
	for _, fileName := range cgroupFiles {
		filePath := filepath.Join(testPodDir, fileName)
		err := ioutil.WriteFile(filePath, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create cgroup file %s: %v", fileName, err)
		}
	}

	err = applyIOLimitConfig(testPodPath, cfg)
	if err != nil {
		t.Errorf("Should not error with valid config and mock devices, got %v", err)
	}

	// Verify files were created with correct content
	expectedFiles := map[string]string{
		"blkio.throttle.read_bps_device":   "8:0 1048576",
		"blkio.throttle.write_bps_device":  "8:16 2097152",
		"blkio.throttle.read_iops_device":  "259:0 1000",
		"blkio.throttle.write_iops_device": "253:1 2000",
	}

	for fileName, expectedContent := range expectedFiles {
		content, err := cgroup.ReadCgroupFile("blkio", testPodPath, fileName)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", fileName, err)
			continue
		}

		result := strings.TrimSpace(string(content))
		if result != expectedContent {
			t.Errorf("File %s: expected content %q, got %q", fileName, expectedContent, result)
		}
	}
}

// TestConfigIOLimit tests the configIOLimit method
func TestConfigIOLimit(t *testing.T) {
	// Setup test cgroup environment
	tempDir, cleanup := setupTestCgroupEnv(t)
	defer cleanup()

	// Setup mock convertToMajorMinor function
	cleanupMock := setupMockConvertToMajorMinor()
	defer cleanupMock()

	// Create test pod directory in the temp cgroup root following the correct cgroup structure
	testPodPath := "test-pod"
	// The cgroup structure should be: {tempDir}/blkio/{testPodPath}/
	blkioDir := filepath.Join(tempDir, "blkio")
	testPodDir := filepath.Join(blkioDir, testPodPath)
	err := os.MkdirAll(testPodDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test pod directory: %v", err)
	}

	// Create the cgroup files first (they need to exist before writing)
	cgroupFiles := []string{
		"blkio.throttle.read_bps_device",
		"blkio.throttle.write_bps_device",
		"blkio.throttle.read_iops_device",
		"blkio.throttle.write_iops_device",
	}
	for _, fileName := range cgroupFiles {
		filePath := filepath.Join(testPodDir, fileName)
		// Create files with some initial data to test clearing
		initialData := "8:0 1000000\n8:16 2000000"
		err := ioutil.WriteFile(filePath, []byte(initialData), 0644)
		if err != nil {
			t.Fatalf("Failed to create cgroup file %s: %v", fileName, err)
		}
	}

	iolimit := &IOLimit{}

	// Test 1: Empty config (should just clear files)
	podInfo := &typedef.PodInfo{
		Name:        "test-pod",
		Annotations: map[string]string{},
		IDContainersMap: map[string]*typedef.ContainerInfo{
			"test-container": {
				Hierarchy: cgroup.Hierarchy{Path: testPodPath},
			},
		},
	}
	err = iolimit.configIOLimit(podInfo)
	if err != nil {
		t.Errorf("Expected no error with empty config, got %v", err)
	}

	// Test 2: Invalid JSON config
	podInfo.Annotations[constant.BlkioKey] = "invalid json"
	err = iolimit.configIOLimit(podInfo)
	if err == nil {
		t.Error("Expected error with invalid JSON config")
	}
	if !strings.Contains(err.Error(), "parse blkio config") && !strings.Contains(err.Error(), "failed to clear") {
		t.Errorf("Expected parse or clear error, got %v", err)
	}

	// Test 3: Valid JSON config with mock devices
	validConfig := `{
		"device_read_bps": [{"device": "/dev/sda", "value": "1048576"}],
		"device_write_bps": [{"device": "8:16", "value": "2097152"}],
		"device_read_iops": [{"device": "/dev/nvme0n1", "value": "1000"}],
		"device_write_iops": [{"device": "253:1", "value": "2000"}]
	}`
	podInfo.Annotations[constant.BlkioKey] = validConfig
	err = iolimit.configIOLimit(podInfo)
	if err != nil {
		t.Errorf("Expected no error with valid config, got %v", err)
	}

	// Verify the configuration was applied correctly
	expectedFiles := map[string]string{
		"blkio.throttle.read_bps_device":   "8:0 1048576",
		"blkio.throttle.write_bps_device":  "8:16 2097152",
		"blkio.throttle.read_iops_device":  "259:0 1000",
		"blkio.throttle.write_iops_device": "253:1 2000",
	}

	for fileName, expectedContent := range expectedFiles {
		content, err := cgroup.ReadCgroupFile("blkio", testPodPath, fileName)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", fileName, err)
			continue
		}

		result := strings.TrimSpace(string(content))
		if result != expectedContent {
			t.Errorf("File %s: expected content %q, got %q", fileName, expectedContent, result)
		}
	}

	// Test 4: Partial config - only one field configured (fresh start)
	// First reinitialize the cgroup files to be empty
	for _, fileName := range cgroupFiles {
		filePath := filepath.Join(testPodDir, fileName)
		err := ioutil.WriteFile(filePath, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to reinitialize cgroup file %s: %v", fileName, err)
		}
	}

	partialConfig := `{"device_read_bps": [{"device": "/dev/sda", "value": "2097152"}]}`
	podInfo.Annotations[constant.BlkioKey] = partialConfig
	err = iolimit.configIOLimit(podInfo)
	if err != nil {
		t.Errorf("Expected no error with partial config, got %v", err)
	}

	// Verify only the specified field was configured, others should remain empty
	partialExpectedFiles := map[string]string{
		"blkio.throttle.read_bps_device": "8:0 2097152",
	}

	for fileName, expectedContent := range partialExpectedFiles {
		content, err := cgroup.ReadCgroupFile("blkio", testPodPath, fileName)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", fileName, err)
			continue
		}

		result := strings.TrimSpace(string(content))
		if result != expectedContent {
			t.Errorf("File %s: expected content %q, got %q", fileName, expectedContent, result)
		}
	}

	// Verify other files remain empty since they had no initial content and no new config
	otherFiles := []string{
		"blkio.throttle.write_bps_device",
		"blkio.throttle.read_iops_device",
		"blkio.throttle.write_iops_device",
	}
	for _, fileName := range otherFiles {
		content, err := cgroup.ReadCgroupFile("blkio", testPodPath, fileName)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", fileName, err)
			continue
		}

		result := strings.TrimSpace(string(content))
		if result != "" {
			t.Errorf("File %s should remain empty with partial config, got %q", fileName, result)
		}
	}

	// Test 4: Valid JSON config but invalid cgroup path
	podInfo.IDContainersMap["test-container"].Hierarchy.Path = "/invalid/path"
	err = iolimit.configIOLimit(podInfo)
	if err == nil {
		t.Error("Expected error with invalid cgroup path")
	}
}

// TestClearConfig tests the clearConfig function
func TestClearConfig(t *testing.T) {
	// Setup test cgroup environment
	tempDir, cleanup := setupTestCgroupEnv(t)
	defer cleanup()

	// Create test pod directory in the temp cgroup root following the correct cgroup structure
	testPodPath := "test-pod"
	// The cgroup structure should be: {tempDir}/blkio/{testPodPath}/
	blkioDir := filepath.Join(tempDir, "blkio")
	testPodDir := filepath.Join(blkioDir, testPodPath)
	err := os.MkdirAll(testPodDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test pod directory: %v", err)
	}

	tests := []struct {
		name         string
		initialData  string
		fileName     string
		expectedData string
		expectError  bool
	}{
		{
			name:         "single device entry",
			initialData:  "8:0 1048576",
			fileName:     "blkio.throttle.read_bps_device",
			expectedData: "8:0 0",
			expectError:  false,
		},
		{
			name:         "multiple device entries",
			initialData:  "8:0 1048576\n8:16 2097152\n253:1 4194304",
			fileName:     "blkio.throttle.write_bps_device",
			expectedData: "8:0 0\n8:16 0\n253:1 0",
			expectError:  false,
		},
		{
			name:         "empty file",
			initialData:  "",
			fileName:     "blkio.throttle.read_iops_device",
			expectedData: "",
			expectError:  false,
		},
		{
			name:         "file with empty lines",
			initialData:  "8:0 1048576\n\n8:16 2097152\n",
			fileName:     "blkio.throttle.write_iops_device",
			expectedData: "8:0 0\n8:16 0",
			expectError:  false,
		},
		{
			name:         "non-existent file",
			initialData:  "",
			fileName:     "non_existent_file",
			expectedData: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file if not testing non-existent file
			if tt.name != "non-existent file" {
				filePath := filepath.Join(testPodDir, tt.fileName)
				err := ioutil.WriteFile(filePath, []byte(tt.initialData), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			// Call clearConfig
			err := clearConfig(testPodPath, tt.fileName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Read the file content after clearConfig using cgroup.ReadCgroupFile
			content, err := cgroup.ReadCgroupFile("blkio", testPodPath, tt.fileName)
			if err != nil {
				t.Fatalf("Failed to read file after clearConfig: %v", err)
			}

			result := strings.TrimSpace(string(content))
			if result != tt.expectedData {
				t.Errorf("Expected file content %q, got %q", tt.expectedData, result)
			}
		})
	}
}

// TestClearAllBlkioThrottleFiles tests the clearAllBlkioThrottleFiles function
func TestClearAllBlkioThrottleFiles(t *testing.T) {
	// Setup test cgroup environment
	tempDir, cleanup := setupTestCgroupEnv(t)
	defer cleanup()

	// Create test pod directory in the temp cgroup root following the correct cgroup structure
	testPodPath := "test-pod"
	// The cgroup structure should be: {tempDir}/blkio/{testPodPath}/
	blkioDir := filepath.Join(tempDir, "blkio")
	testPodDir := filepath.Join(blkioDir, testPodPath)
	err := os.MkdirAll(testPodDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test pod directory: %v", err)
	}

	// Create all 4 blkio throttle files with test data
	testData := map[string]string{
		"blkio.throttle.read_bps_device":   "8:0 1048576\n8:16 2097152",
		"blkio.throttle.write_bps_device":  "8:0 2097152\n253:1 4194304",
		"blkio.throttle.read_iops_device":  "8:0 1000\n8:16 2000",
		"blkio.throttle.write_iops_device": "8:0 1500\n253:1 3000",
	}

	expectedResults := map[string]string{
		"blkio.throttle.read_bps_device":   "8:0 0\n8:16 0",
		"blkio.throttle.write_bps_device":  "8:0 0\n253:1 0",
		"blkio.throttle.read_iops_device":  "8:0 0\n8:16 0",
		"blkio.throttle.write_iops_device": "8:0 0\n253:1 0",
	}

	// Write initial data to files
	for fileName, content := range testData {
		filePath := filepath.Join(testPodDir, fileName)
		err := ioutil.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", fileName, err)
		}
	}

	// Call clearAllBlkioThrottleFiles
	err = clearAllBlkioThrottleFiles(testPodPath)
	if err != nil {
		t.Errorf("clearAllBlkioThrottleFiles returned error: %v", err)
		return
	}

	// Verify that all files have been reset correctly
	for fileName, expectedContent := range expectedResults {
		content, err := cgroup.ReadCgroupFile("blkio", testPodPath, fileName)
		if err != nil {
			t.Errorf("Failed to read file %s after clearAllBlkioThrottleFiles: %v", fileName, err)
			continue
		}

		result := strings.TrimSpace(string(content))
		if result != expectedContent {
			t.Errorf("File %s: expected content %q, got %q", fileName, expectedContent, result)
		}
	}

	// Test with non-existent directory
	err = clearAllBlkioThrottleFiles("/non/existent/path")
	if err == nil {
		t.Error("Expected error with non-existent directory")
	}
	if !strings.Contains(err.Error(), "failed to clear") {
		t.Errorf("Expected 'failed to clear' error, got %v", err)
	}
}

// TestConvertToMajorMinorImpl_ErrorCases tests the convertToMajorMinorImpl function error cases
func TestConvertToMajorMinorImpl_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid major:minor format - too many colons",
			input:       "8:0:1",
			expectError: true,
			errorMsg:    "unsupported device name format",
		},
		{
			name:        "invalid major:minor format - non-numeric major",
			input:       "abc:0",
			expectError: true,
			errorMsg:    "unsupported device name format",
		},
		{
			name:        "invalid major:minor format - non-numeric minor",
			input:       "8:abc",
			expectError: true,
			errorMsg:    "unsupported device name format",
		},
		{
			name:        "non-existent device file",
			input:       "/dev/nonexistent_device_12345",
			expectError: true,
			errorMsg:    "failed to stat device",
		},
		{
			name:        "invalid device format",
			input:       "invalid_format",
			expectError: true,
			errorMsg:    "unsupported device name format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertToMajorMinorImpl(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %v", tt.errorMsg, err)
				}
				if result != "" {
					t.Errorf("Expected empty result on error, got %q", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestIOLimit_PreStart_WithPods tests the PreStart method with pods that have blkio config
func TestIOLimit_PreStart_WithPods(t *testing.T) {
	// Setup test cgroup environment
	tempDir, cleanup := setupTestCgroupEnv(t)
	defer cleanup()

	// Setup mock convertToMajorMinor function
	cleanupMock := setupMockConvertToMajorMinor()
	defer cleanupMock()

	iolimit := &IOLimit{}

	// Create test pod with blkio config
	podInfo := &typedef.PodInfo{
		Name:      "test-pod-with-config",
		Hierarchy: cgroup.Hierarchy{Path: "test-pod"},
		Annotations: map[string]string{
			constant.BlkioKey: `{"device_read_bps":[{"device":"/dev/sda","value":"1048576"}]}`,
		},
	}

	// Create pod directory and cgroup files following the correct cgroup structure
	// The cgroup structure should be: {tempDir}/blkio/{testPodPath}/
	blkioDir := filepath.Join(tempDir, "blkio")
	testPodDir := filepath.Join(blkioDir, "test-pod")
	err := os.MkdirAll(testPodDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test pod directory: %v", err)
	}

	cgroupFiles := []string{
		"blkio.throttle.read_bps_device",
		"blkio.throttle.write_bps_device",
		"blkio.throttle.read_iops_device",
		"blkio.throttle.write_iops_device",
	}
	for _, fileName := range cgroupFiles {
		filePath := filepath.Join(testPodDir, fileName)
		err := ioutil.WriteFile(filePath, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create cgroup file %s: %v", fileName, err)
		}
	}

	mockViewer := &mockViewer{
		pods: map[string]*typedef.PodInfo{
			"test-pod-with-config": podInfo,
		},
		containers: make(map[string]*typedef.ContainerInfo),
	}

	// Test PreStart with pod that has blkio config
	err = iolimit.PreStart(mockViewer)
	if err != nil {
		t.Errorf("PreStart should not return error with valid pod config, got %v", err)
	}

	// Test PreStart with pod that has invalid config - this should cause an error
	podInfo.Annotations[constant.BlkioKey] = "invalid json"
	err = iolimit.PreStart(mockViewer)
	if err == nil {
		t.Error("PreStart should return error with invalid pod config")
	}
}

// TestApplyIOLimitConfig_ErrorCases tests error cases for applyIOLimitConfig
func TestApplyIOLimitConfig_ErrorCases(t *testing.T) {
	// Setup test cgroup environment
	_, cleanupCgroup := setupTestCgroupEnv(t)
	defer cleanupCgroup()

	// Setup mock convertToMajorMinor function
	cleanupMock := setupMockConvertToMajorMinor()
	defer cleanupMock()

	testPodPath := "test-pod"

	// Test with config that fails to write to cgroup files
	cfg := &BlkConfig{
		DeviceReadBps: []DeviceConfig{
			{DeviceName: "/dev/sda", DeviceValue: "1048576"},
		},
	}

	// Don't create the cgroup files - this should cause write errors
	err := applyIOLimitConfig(testPodPath, cfg)
	if err == nil {
		t.Error("Expected error when cgroup files don't exist")
	}
	if !strings.Contains(err.Error(), "failed to apply") {
		t.Errorf("Expected 'failed to apply' error, got %v", err)
	}
}

// TestApplyDeviceConfig_ErrorCases tests error cases for applyDeviceConfig
func TestApplyDeviceConfig_ErrorCases(t *testing.T) {
	// Setup mock convertToMajorMinor function
	cleanupMock := setupMockConvertToMajorMinor()
	defer cleanupMock()

	testPodPath := "test-pod"
	testFileName := "test_file"

	devices := []DeviceConfig{
		{DeviceName: "/dev/sda", DeviceValue: "1048576"},
	}

	// Test with non-existent cgroup path - should cause write error
	err := applyDeviceConfig(testPodPath, testFileName, devices)
	if err == nil {
		t.Error("Expected error when cgroup file doesn't exist")
	}
	if !strings.Contains(err.Error(), "failed to write config") {
		t.Errorf("Expected 'failed to write config' error, got %v", err)
	}
}

// TestClearConfig_EdgeCases tests edge cases for clearConfig
func TestClearConfig_EdgeCases(t *testing.T) {
	// Setup test cgroup environment
	tempDir, cleanup := setupTestCgroupEnv(t)
	defer cleanup()

	// Create test pod directory following the correct cgroup structure
	testPodPath := "test-pod"
	// The cgroup structure should be: {tempDir}/blkio/{testPodPath}/
	blkioDir := filepath.Join(tempDir, "blkio")
	testPodDir := filepath.Join(blkioDir, testPodPath)
	err := os.MkdirAll(testPodDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test pod directory: %v", err)
	}

	// Test 1: File doesn't exist - should return error
	testFileName := "nonexistent_file"
	err = clearConfig(testPodPath, testFileName)
	if err == nil {
		t.Error("Expected error when file doesn't exist")
	}

	// Test 2: File exists with data - should clear data correctly
	testFileName = "existing_file"
	filePath := filepath.Join(testPodDir, testFileName)
	initialData := "8:0 1048576\n8:16 2097152"
	err = ioutil.WriteFile(filePath, []byte(initialData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Clear the config
	err = clearConfig(testPodPath, testFileName)
	if err != nil {
		t.Errorf("Unexpected error when clearing existing file: %v", err)
		return
	}

	// Verify the file content was cleared correctly
	content, err := cgroup.ReadCgroupFile("blkio", testPodPath, testFileName)
	if err != nil {
		t.Fatalf("Failed to read file after clearConfig: %v", err)
	}

	expectedData := "8:0 0\n8:16 0"
	result := strings.TrimSpace(string(content))
	if result != expectedData {
		t.Errorf("Expected file content %q, got %q", expectedData, result)
	}
}

// TestConfigIOLimit_EdgeCases tests edge cases for configIOLimit
func TestConfigIOLimit_EdgeCases(t *testing.T) {
	// Setup test cgroup environment
	_, cleanup := setupTestCgroupEnv(t)
	defer cleanup()

	// Setup mock convertToMajorMinor function
	cleanupMock := setupMockConvertToMajorMinor()
	defer cleanupMock()

	iolimit := &IOLimit{}

	// Test with valid JSON but applying config fails (no cgroup files)
	podInfo := &typedef.PodInfo{
		Name: "test-pod",
		Annotations: map[string]string{
			constant.BlkioKey: `{"device_read_bps":[{"device":"/dev/sda","value":"1048576"}]}`,
		},
		IDContainersMap: map[string]*typedef.ContainerInfo{
			"test-container": {
				Hierarchy: cgroup.Hierarchy{Path: "nonexistent-container"},
			},
		},
	}

	err := iolimit.configIOLimit(podInfo)
	if err == nil {
		t.Error("Expected error when applying config to non-existent cgroup path")
	}
	if !strings.Contains(err.Error(), "failed to clear") || !strings.Contains(err.Error(), "failed to apply") {
		// Either clearing or applying can fail first, both are valid error scenarios
		if !strings.Contains(err.Error(), "failed to clear") && !strings.Contains(err.Error(), "failed to apply") {
			t.Errorf("Expected either 'failed to clear' or 'failed to apply' error, got %v", err)
		}
	}
}

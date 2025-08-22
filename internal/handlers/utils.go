package handlers

import (
	"fmt"
	"strconv"
)

// parseIntParam 解析整数参数
func parseIntParam(value, paramName string) (int, error) {
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: %s", paramName, value)
	}
	return result, nil
}

// parseUint64Param 解析uint64参数
func parseUint64Param(value, paramName string) (uint64, error) {
	result, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: %s", paramName, value)
	}
	return result, nil
}

// parseFloat64Param 解析float64参数
func parseFloat64Param(value, paramName string) (float64, error) {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: %s", paramName, value)
	}
	return result, nil
}

// parseBoolParam 解析布尔参数
func parseBoolParam(value, paramName string) (bool, error) {
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid %s parameter: %s", paramName, value)
	}
	return result, nil
}

// validateRequired 验证必需参数
func validateRequired(value, paramName string) error {
	if value == "" {
		return fmt.Errorf("%s is required", paramName)
	}
	return nil
}

// validateRange 验证数值范围
func validateRange(value, min, max int, paramName string) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d", paramName, min, max)
	}
	return nil
}

// validateUint64Range 验证uint64数值范围
func validateUint64Range(value, min, max uint64, paramName string) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %d and %d", paramName, min, max)
	}
	return nil
}

// validateFloat64Range 验证float64数值范围
func validateFloat64Range(value, min, max float64, paramName string) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %f and %f", paramName, min, max)
	}
	return nil
}
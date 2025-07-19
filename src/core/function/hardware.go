package function

import (
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// HardwareController 硬件控制器接口
type HardwareController interface {
	// 电机控制
	SetMotorSpeed(motorID int, speed int) error
	SetMotorDirection(motorID int, direction string) error

	// LED控制
	SetLED(ledID int, state bool) error
	SetLEDColor(ledID int, r, g, b int) error

	// 传感器读取
	GetTemperature() (float64, error)
	GetHumidity() (float64, error)
	GetDistance() (float64, error)
}

// 全局硬件控制器实例
var GlobalHardwareController HardwareController

// SetHardwareController 设置硬件控制器
func SetHardwareController(controller HardwareController) {
	GlobalHardwareController = controller
}

// RegisterHardwareFunctions 注册硬件控制函数
func (fr *FunctionRegistry) RegisterHardwareFunctions() error {
	// 电机控制函数
	motorSpeedFunc := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "set_motor_speed",
			Description: "控制电机转速",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"motor_id": map[string]interface{}{
						"type":        "integer",
						"description": "电机ID (1-4)",
					},
					"speed": map[string]interface{}{
						"type":        "integer",
						"description": "转速 (-100 到 100, 负数表示反转)",
					},
				},
				"required": []string{"motor_id", "speed"},
			},
		},
	}

	motorDirectionFunc := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "set_motor_direction",
			Description: "设置电机方向",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"motor_id": map[string]interface{}{
						"type":        "integer",
						"description": "电机ID (1-4)",
					},
					"direction": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"forward", "backward", "stop"},
						"description": "电机方向",
					},
				},
				"required": []string{"motor_id", "direction"},
			},
		},
	}

	// LED控制函数
	ledControlFunc := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "set_led",
			Description: "控制LED开关",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"led_id": map[string]interface{}{
						"type":        "integer",
						"description": "LED ID (1-8)",
					},
					"state": map[string]interface{}{
						"type":        "boolean",
						"description": "LED状态 (true=开, false=关)",
					},
				},
				"required": []string{"led_id", "state"},
			},
		},
	}

	ledColorFunc := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "set_led_color",
			Description: "设置LED颜色 (RGB)",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"led_id": map[string]interface{}{
						"type":        "integer",
						"description": "LED ID (1-8)",
					},
					"r": map[string]interface{}{
						"type":        "integer",
						"description": "红色分量 (0-255)",
					},
					"g": map[string]interface{}{
						"type":        "integer",
						"description": "绿色分量 (0-255)",
					},
					"b": map[string]interface{}{
						"type":        "integer",
						"description": "蓝色分量 (0-255)",
					},
				},
				"required": []string{"led_id", "r", "g", "b"},
			},
		},
	}

	// 传感器读取函数
	temperatureFunc := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_temperature",
			Description: "获取温度传感器数据",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	humidityFunc := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_humidity",
			Description: "获取湿度传感器数据",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	distanceFunc := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_distance",
			Description: "获取距离传感器数据",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	// 注册所有函数
	functions := []openai.Tool{
		motorSpeedFunc,
		motorDirectionFunc,
		ledControlFunc,
		ledColorFunc,
		temperatureFunc,
		humidityFunc,
		distanceFunc,
	}

	for _, function := range functions {
		if err := fr.RegisterFunction(function.Function.Name, function); err != nil {
			return fmt.Errorf("注册硬件函数失败: %v", err)
		}
	}

	return nil
}

// HandleHardwareFunction 处理硬件控制函数调用
func HandleHardwareFunction(functionName string, arguments json.RawMessage) (interface{}, error) {
	if GlobalHardwareController == nil {
		return nil, fmt.Errorf("硬件控制器未初始化")
	}

	switch functionName {
	case "set_motor_speed":
		var args struct {
			MotorID int `json:"motor_id"`
			Speed   int `json:"speed"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, fmt.Errorf("解析参数失败: %v", err)
		}
		err := GlobalHardwareController.SetMotorSpeed(args.MotorID, args.Speed)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("电机 %d 转速设置为 %d", args.MotorID, args.Speed),
		}, nil

	case "set_motor_direction":
		var args struct {
			MotorID   int    `json:"motor_id"`
			Direction string `json:"direction"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, fmt.Errorf("解析参数失败: %v", err)
		}
		err := GlobalHardwareController.SetMotorDirection(args.MotorID, args.Direction)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("电机 %d 方向设置为 %s", args.MotorID, args.Direction),
		}, nil

	case "set_led":
		var args struct {
			LEDID int  `json:"led_id"`
			State bool `json:"state"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, fmt.Errorf("解析参数失败: %v", err)
		}
		err := GlobalHardwareController.SetLED(args.LEDID, args.State)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("LED %d 状态设置为 %v", args.LEDID, args.State),
		}, nil

	case "set_led_color":
		var args struct {
			LEDID int `json:"led_id"`
			R     int `json:"r"`
			G     int `json:"g"`
			B     int `json:"b"`
		}
		if err := json.Unmarshal(arguments, &args); err != nil {
			return nil, fmt.Errorf("解析参数失败: %v", err)
		}
		err := GlobalHardwareController.SetLEDColor(args.LEDID, args.R, args.G, args.B)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("LED %d 颜色设置为 RGB(%d,%d,%d)", args.LEDID, args.R, args.G, args.B),
		}, nil

	case "get_temperature":
		temp, err := GlobalHardwareController.GetTemperature()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"temperature": temp,
			"unit":        "°C",
		}, nil

	case "get_humidity":
		humidity, err := GlobalHardwareController.GetHumidity()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"humidity": humidity,
			"unit":     "%",
		}, nil

	case "get_distance":
		distance, err := GlobalHardwareController.GetDistance()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"distance": distance,
			"unit":     "cm",
		}, nil

	default:
		return nil, fmt.Errorf("未知的硬件函数: %s", functionName)
	}
}

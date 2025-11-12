package validation

import (
	"context"
	"reflect"
	"strings"

	"k8s.io/klog/v2"
)

type Validator struct {
	registry map[string]reflect.Value
}

// NewValidator 构造自定义验证器
func NewValidator(customValidator any) *Validator {
	return &Validator{registry: extractValidationMethods(customValidator)}
}

// Validate 使用对应的验证函数验证请求
func (v *Validator) Validate(ctx context.Context, request any) error {
	validationFunc, ok := v.registry[reflect.TypeOf(request).Elem().Name()]
	if !ok {
		// 没有对应的验证函数
		return nil
	}

	result := validationFunc.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(request)})
	if !result[0].IsNil() {
		return result[0].Interface().(error)
	}

	return nil
}

// extractValidationMethods 返回一个值为验证函数的map
// 验证函数从自定义验证器提取
func extractValidationMethods(customValidator any) map[string]reflect.Value {
	funcs := make(map[string]reflect.Value)
	validatorType := reflect.TypeOf(customValidator)
	validatorValue := reflect.ValueOf(customValidator)

	for i := 0; i < validatorType.NumMethod(); i++ {
		method := validatorType.Method(i)
		methodValue := validatorValue.MethodByName(method.Name)

		// 方法有效且以Validate开头
		if !methodValue.IsValid() || !strings.HasPrefix(method.Name, "Validate") {
			continue
		}

		methodType := methodValue.Type()

		// 确保入参是 context.Context 和1个指针
		if methodType.NumIn() != 2 || methodType.NumOut() != 1 ||
			methodType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() ||
			methodType.In(1).Kind() != reflect.Pointer {
			continue
		}

		// 确保方法名称符合预期的命名约定
		requestTypeName := methodType.In(1).Elem().Name()
		if method.Name != ("Validate" + requestTypeName) {
			continue
		}

		// 确保返回类型是 error
		if methodType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		klog.V(4).InfoS("Registering validator", "validator", requestTypeName)
		funcs[requestTypeName] = methodValue
	}

	return funcs
}
